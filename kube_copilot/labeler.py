from langchain_community.callbacks.streamlit.streamlit_callback_handler import ToolRecord
from langchain_community.callbacks.streamlit.streamlit_callback_handler import LLMThoughtLabeler

CHECKMARK_EMOJI = "✅"
THINKING_EMOJI = ":thinking_face:"
HISTORY_EMOJI = ":books:"
EXCEPTION_EMOJI = "⚠️"


class CustomLLMThoughtLabeler(LLMThoughtLabeler):
    def get_tool_label(self, tool: ToolRecord, is_complete: bool) -> str:
        """Return the label for an LLMThought that has an associated
        tool.

        Parameters
        ----------
        tool
            The tool's ToolRecord

        is_complete
            True if the thought is complete; False if the thought
            is still receiving input.

        Returns
        -------
        The markdown label for the thought's container.

        """
        input = tool.input_str.strip()
        name = tool.name
        emoji = CHECKMARK_EMOJI if is_complete else THINKING_EMOJI
        if name == "_Exception":
            emoji = EXCEPTION_EMOJI
            name = "Parsing error"
        # idx = min([60, len(input)])
        # input = input[0:idx]
        # if len(tool.input_str) > idx:
        #     input = input + "..."
        # input = input.replace("\n", " ")
        label = f"{emoji} **{name}:** {input}"
        return label
