# -*- coding: utf-8 -*-
import logging
import os
import sys

import streamlit as st
import yaml
from langchain.callbacks import StreamlitCallbackHandler

from kube_copilot.chains import ReActLLM
from kube_copilot.prompts import get_generate_prompt
from kube_copilot.shell import KubeProcess
from kube_copilot.labeler import CustomLLMThoughtLabeler

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
if st.button("Generate", key="generate"):
    if not os.getenv("OPENAI_API_KEY", ""):
        if not openai_api_key:
            st.info("Please add your OpenAI API key to continue.")
            st.stop()

        os.environ["OPENAI_API_KEY"] = openai_api_key
        os.environ["OPENAI_API_BASE"] = openai_api_base
        os.environ["GOOGLE_API_KEY"] = google_api_key
        os.environ["GOOGLE_CSE_ID"] = google_cse_id


    if not prompt:
        st.info("Please add your prompt to continue.")
        st.stop()

    st.session_state["response"] = ""
    st.session_state["manifests"] = ""
    st_cb = StreamlitCallbackHandler(st.container(), thought_labeler=CustomLLMThoughtLabeler())
    chain = ReActLLM(model=model,
                     verbose=True,
                     enable_python=True,
                     auto_approve=True)
    response = chain.run(get_generate_prompt(prompt), callbacks=[st_cb])
    st.session_state["response"] = response

if st.session_state.get("response", "") != "":
    response = st.session_state.get("response", "")
    with st.container():
        st.markdown(response)

    manifests = response.removeprefix(
        '```').removeprefix('yaml').removesuffix('```').strip()

    try:
        yamls = yaml.safe_load_all(manifests)
        st.session_state["manifests"] = manifests
    except Exception as e:
        st.error("The generated manifests are not valid YAML.")
        st.stop()

if st.session_state.get("manifests", "") != "":
    if st.button("Apply to the cluster", key="apply_manifests"):
        st.write("Applying the generated manifests...")
        st.write(KubeProcess(command="kubectl").run(
            'kubectl apply -f -', input=bytes(manifests, 'utf-8')))
