# -*- coding: utf-8 -*-
import os
from langchain.chat_models import ChatOpenAI
from langchain.agents import AgentType, Tool, initialize_agent
from langchain.agents.agent import AgentExecutor
from langchain.agents.structured_chat.base import StructuredChatAgent
from langchain.utilities import GoogleSearchAPIWrapper
from langchain.experimental.plan_and_execute import PlanAndExecute, load_chat_planner
from langchain.experimental.plan_and_execute.executors.base import ChainExecutor
from kube_copilot.shell import KubeProcess
from kube_copilot.prompts import get_planner_prompt
from langchain.agents.structured_chat.base import StructuredChatAgent
from langchain.callbacks import StdOutCallbackHandler


HUMAN_MESSAGE_TEMPLATE = """Previous steps: {previous_steps}

Current objective: {current_step}

{agent_scratchpad}"""


class PlanAndExecuteLLM:
    '''Wrapper for LLM chain.'''

    def __init__(self, verbose=True, model="gpt-4", additional_tools=None):
        '''Initialize the LLM chain.'''
        self.chain = self.get_chain(
            verbose, model, additional_tools=additional_tools)

    def run(self, instructions):
        '''Run the LLM chain.'''
        return self.chain.run(instructions)

    def get_chain(self, verbose=True, model="gpt-4", additional_tools=None):
        '''Initialize the LLM chain with useful tools.'''
        llm, tools = get_llm_tools(model, additional_tools)

        # executor = load_agent_executor(llm, tools, verbose=verbose)
        agent = StructuredChatAgent.from_llm_and_tools(
            llm,
            tools,
            human_message_template=HUMAN_MESSAGE_TEMPLATE,
            input_variables=["previous_steps",
                             "current_step",
                             "agent_scratchpad"],
            # TODO: Workaround for issue https://github.com/hwchase17/langchain/issues/1358.
            handle_parsing_errors="Check your output and make sure it conforms!",
        )

        agent_executor = AgentExecutor.from_agent_and_tools(
            agent=agent, tools=tools, verbose=verbose
        )
        executor = ChainExecutor(chain=agent_executor)
        planner = load_chat_planner(
            llm=llm, system_prompt=get_planner_prompt())
        step_handler = StdOutCallbackHandler(color="green")
        return PlanAndExecute(planner=planner, executor=executor, verbose=verbose, callbacks=[step_handler])


class ReActLLM:
    '''Wrapper for LLM chain.'''

    def __init__(self, verbose=True, model="gpt-4", additional_tools=None):
        '''Initialize the LLM chain.'''
        self.chain = self.get_chain(
            verbose, model, additional_tools=additional_tools)

    def run(self, instructions):
        '''Run the LLM chain.'''
        return self.chain.run(instructions)

    def get_chain(self, verbose=True, model="gpt-4", additional_tools=None):
        '''Initialize the LLM chain with useful tools.'''
        llm, tools = get_llm_tools(model, additional_tools)
        agent = initialize_agent(llm=llm,
                                 tools=tools,
                                 verbose=verbose,
                                 type=AgentType.CHAT_ZERO_SHOT_REACT_DESCRIPTION,
                                 handle_parsing_errors=handle_parsing_error)
        return agent


def get_llm_tools(model, additional_tools):
    '''Initialize the LLM chain with useful tools.'''
    if os.getenv("OPENAI_API_TYPE") == "azure" or (os.getenv("OPENAI_API_BASE") is not None and "azure" in os.getenv("OPENAI_API_BASE")):
        engine = model.replace(".", "")
        llm = ChatOpenAI(model_name=model,
                         temperature=0,
                         request_timeout=120,
                         model_kwargs={"engine": engine})
    else:
        llm = ChatOpenAI(model_name=model,
                         temperature=0,
                         request_timeout=120)

    tools = [
        Tool(
            name="kubectl",
            description="Useful for executing kubectl command to query information from kubernetes cluster. Input: a kubectl get command. Output: the yaml for the resource.",
            func=KubeProcess(command="kubectl").run,
        ),
        Tool(
            name="trivy",
            description="Useful for executing trivy image command to scan images for vulnerabilities. Input: a trivy image command. Output: the vulnerabilities found in the image.",
            func=KubeProcess(command="trivy").run,
        ),
    ]

    if os.getenv("GOOGLE_API_KEY") and os.getenv("GOOGLE_CSE_ID"):
        tools += [
            Tool(
                name="search",
                func=GoogleSearchAPIWrapper(
                    google_api_key=os.getenv("GOOGLE_API_KEY"),
                    google_cse_id=os.getenv("GOOGLE_CSE_ID"),
                ).run,
                description="search the web for current events or current state of the world"
            )
        ]

    if additional_tools is not None:
        tools += additional_tools
    return llm, tools


def handle_parsing_error(error) -> str:
    '''Helper function to handle parsing errors from LLM.'''
    return str(error).split("Could not parse LLM output:")[1].strip().removeprefix('`').removesuffix('`')
