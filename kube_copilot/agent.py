# -*- coding: utf-8 -*-
import os
from langchain.chat_models import ChatOpenAI
from langchain.agents import Tool, load_tools
from langchain.memory import ConversationBufferMemory
from langchain.agents import initialize_agent
from langchain.tools.python.tool import PythonREPLTool
from langchain.utilities import GoogleSearchAPIWrapper


class CopilotLLM:

    def __init__(self, verbose=True, additional_tools=[]):
        self.chain = get_chat_chain(verbose, additional_tools)

    def run(self, instructions):
        try:
            result = self.chain.run(instructions)
            return result
        except Exception as e:
            # Workaround for issue https://github.com/hwchase17/langchain/issues/1358.
            if "Could not parse LLM output:" in str(e):
                return str(e).split("Could not parse LLM output:")[1]
            else:
                raise e


def get_chat_chain(verbose=True, additional_tools=[], agent="chat-zero-shot-react-description", max_iterations=30):
    '''Initialize the LLM chain with useful tools.'''
    llm = ChatOpenAI(engine="gpt-35-turbo", max_tokens=512)

    tools = load_tools(["terminal", "human"], llm)
    if os.getenv("KUBE_COPILOT_ENABLE_PYTHON"):
        tools += [
            Tool(
                name="Python",
                func=PythonREPLTool().run,
                description="helps to run Python codes"
            )
        ]

    if os.getenv("GOOGLE_API_KEY") != "" and os.getenv("GOOGLE_CSE_ID") != "":
        tools += [
            Tool(
                name="Search",
                func=GoogleSearchAPIWrapper(
                    google_api_key=os.getenv("GOOGLE_API_KEY"),
                    google_cse_id=os.getenv("GOOGLE_CSE_ID"),
                ).run,
                description="search the web for current events or current state of the world"
            )
        ]

    tools += additional_tools

    memory = ConversationBufferMemory(
        memory_key="chat_history", return_messages=True)
    chain = initialize_agent(
        tools, llm, agent=agent, memory=memory,
        verbose=verbose, max_iterations=max_iterations)
    return chain
