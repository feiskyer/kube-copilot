# -*- coding: utf-8 -*-
import os
from typing import List, Union
from langchain.chat_models import ChatOpenAI
from langchain.agents import Tool, load_tools
from langchain.memory import ConversationBufferMemory
from langchain.agents import initialize_agent
from langchain.tools.python.tool import PythonREPLTool
from langchain.utilities import GoogleSearchAPIWrapper
from langchain.utilities.bash import BashProcess


class CopilotLLM:
    '''Wrapper for LLM chain.'''

    def __init__(self, verbose=True, model="gpt-3.5-turbo", additional_tools=None, enable_terminal=False):
        '''Initialize the LLM chain.'''
        self.chain = get_chat_chain(verbose, model, additional_tools=additional_tools,
                                    enable_terminal=enable_terminal)

    def run(self, instructions):
        '''Run the LLM chain.'''
        try:
            result = self.chain.run(instructions)
            return result
        except Exception as e:
            # Workaround for issue https://github.com/hwchase17/langchain/issues/1358.
            if "Could not parse LLM output:" in str(e):
                return str(e).split("Could not parse LLM output:")[1]
            else:
                raise e


class KubeProcess(BashProcess):
    '''Wrapper for kubectl/docker/trivy/helm commands.'''

    supported_commands = ["kubectl", "docker", "trivy", "helm"]

    def run(self, commands: Union[str, List[str]]) -> str:
        if isinstance(commands, str):
            commands = [commands]
        commands = ";".join(commands)
        if not any(commands.startswith(cmd) for cmd in self.supported_commands):
            # raise Exception(f'Commands {commands} is not allowed to run')
            return None

        return super().run(commands)


def get_chat_chain(verbose=True, model="gpt-3.5-turbo", additional_tools=None,
                   agent="chat-zero-shot-react-description",
                   enable_terminal=False, max_iterations=30):
    '''Initialize the LLM chain with useful tools.'''
    if os.getenv("OPENAI_API_TYPE") == "azure":
        if model == "gpt-3.5-turbo":
            model = "gpt-35-turbo"
        llm = ChatOpenAI(engine=model, max_tokens=512, temperature=0)
    else:
        llm = ChatOpenAI(model=model, max_tokens=512, temperature=0)

    tools = load_tools(["human"], llm)
    if enable_terminal:
        tools += load_tools(["terminal"], llm)
    else:
        tools += [
            Tool(
                name="KubeProcess",
                description="Executes kubectl/docker/trivy commands in a terminal. Input should be valid kubectl/docker/trivy commands (other commands are not supported), and the output will be any output from running that command.",
                func=KubeProcess().run,
            )
        ]

    if os.getenv("KUBE_COPILOT_ENABLE_PYTHON"):
        tools += [
            Tool(
                name="Python",
                func=PythonREPLTool().run,
                description="helps to run Python codes"
            )
        ]

    if os.getenv("GOOGLE_API_KEY") and os.getenv("GOOGLE_CSE_ID"):
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

    if additional_tools is not None:
        tools += additional_tools

    memory = ConversationBufferMemory(
        memory_key="chat_history", return_messages=True)
    chain = initialize_agent(
        tools, llm, agent=agent, memory=memory,
        verbose=verbose, max_iterations=max_iterations)
    return chain
