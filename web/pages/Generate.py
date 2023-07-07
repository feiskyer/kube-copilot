# -*- coding: utf-8 -*-
import logging
import os
import sys

import streamlit as st
from langchain.callbacks import StreamlitCallbackHandler

from kube_copilot.chains import ReActLLM
from kube_copilot.llm import init_openai
from kube_copilot.prompts import get_generate_prompt

logging.basicConfig(stream=sys.stdout, level=logging.CRITICAL)
logging.getLogger().addHandler(logging.StreamHandler(stream=sys.stdout))

st.set_page_config(page_title="Generate Kubernetes Manifests", page_icon="💬")
st.title("💬 Generate Kubernetes Manifests")

with st.sidebar:
    model = st.text_input("OpenAI Model",
                          key="openai_api_model", value=os.getenv("OPENAI_API_MODEL", "gpt-4"))

    if not os.getenv("OPENAI_API_KEY", ""):
        # show the setting panel if the API key is not set from environment variable
        openai_api_key = st.text_input(
            "OpenAI API key", key="openai_api_key", type="password", value=os.getenv("OPENAI_API_KEY", ""))
        openai_api_base = st.text_input("OpenAI API base URL", key="openai_api_base", value=os.getenv(
            "OPENAI_API_BASE", "https://api.openai.com/v1"))
        google_api_key = st.text_input(
            "Google API key", key="google_api_key", type="password", value=os.getenv("GOOGLE_API_KEY", ""))
        google_cse_id = st.text_input(
            "Google CSE ID", key="google_cse_id", type="password", value=os.getenv("GOOGLE_CSE_ID", ""))


prompt = st.text_input("Prompt", key="prompt",
                       placeholder="<input prompt here>")
if st.button("Generate"):
    if not openai_api_key:
        st.info("Please add your OpenAI API key to continue.")
        st.stop()

    os.environ["OPENAI_API_KEY"] = openai_api_key
    os.environ["OPENAI_API_BASE"] = openai_api_base
    os.environ["GOOGLE_API_KEY"] = google_api_key
    os.environ["GOOGLE_CSE_ID"] = google_cse_id
    init_openai()

    if not prompt:
        st.info("Please add your prompt to continue.")
        st.stop()

    st_cb = StreamlitCallbackHandler(st.container())
    chain = ReActLLM(model=model,
                     verbose=True,
                     enable_python=True,
                     auto_approve=True)
    response = chain.run(get_generate_prompt(prompt), callbacks=[st_cb])
    st.write(response)

#     # Apply the generated manifests in cluster
#     if click.confirm('Do you approve to apply the generated manifests to cluster?'):
#         manifests = result.removeprefix(
#             '```').removeprefix('yaml').removesuffix('```')
#         print(KubeProcess(command="kubectl").run(
#             'kubectl apply -f -', input=bytes(manifests, 'utf-8')))
