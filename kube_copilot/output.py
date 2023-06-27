import json
from typing import Union

from langchain.agents.agent import AgentOutputParser
from langchain.agents.chat.prompt import FORMAT_INSTRUCTIONS
from langchain.schema import AgentAction, AgentFinish, OutputParserException

FINAL_ANSWER_ACTION = "Final Answer:"


class ChatOutputParser(AgentOutputParser):
    def get_format_instructions(self) -> str:
        return FORMAT_INSTRUCTIONS

    def parse(self, text: str) -> Union[AgentAction, AgentFinish]:
        includes_answer = FINAL_ANSWER_ACTION in text
        try:
            action = text.split("```")[1]
            if action.startswith('python\n'):
                # Ensure the Python code snippets are handled by the Python action.
                response = {
                    "action": "python",
                    "action_input": action.split('python\n')[1],
                }
            elif action.startswith('sh\n') or action.startswith('bash\n'):
                # Ensure the shell code snippets are handled by the kubectl action.
                response = {
                    "action": "kubectl",
                    "action_input": action.split('sh\n')[1],
                }
            else:
                # JSON object is expected by default.
                response = json.loads(action.strip())

            includes_action = "action" in response and "action_input" in response
            if includes_answer and includes_action:
                raise OutputParserException(
                    "Parsing LLM output produced a final answer "
                    f"and a parse-able action: {text}"
                )
            return AgentAction(response["action"], response["action_input"], text)

        except Exception:
            if not includes_answer:
                raise OutputParserException(
                    f"Could not parse LLM output: {text}")
            return AgentFinish(
                {"output": text.split(FINAL_ANSWER_ACTION)[-1].strip()}, text
            )

    @property
    def _type(self) -> str:
        return "chat"
