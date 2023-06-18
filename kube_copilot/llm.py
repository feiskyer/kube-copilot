# -*- coding: utf-8 -*-

import os
import openai


def init_openai():
    """
    Initializes the OpenAI API with the provided API key and sets the API base and version if using Azure.
    """
    if os.getenv("OPENAI_API_KEY") == "":
        raise Exception("Please set OPENAI_API_KEY via environment variable")

    if os.getenv("OPENAI_API_TYPE") == "azure" or (os.getenv("OPENAI_API_BASE") is not None and "azure" in os.getenv("OPENAI_API_BASE")):
        openai.api_type = "azure"
        openai.api_base = os.getenv("OPENAI_API_BASE")
        openai.api_version = "2023-05-15"
        openai.api_key = os.getenv("OPENAI_API_KEY")
    else:
        openai.api_key = os.getenv("OPENAI_API_KEY")
        openai.api_base = "https://api.openai.com/v1"
