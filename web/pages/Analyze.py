# -*- coding: utf-8 -*-
import logging
import os
import sys

import streamlit as st
from langchain_community.callbacks.streamlit.streamlit_callback_handler import StreamlitCallbackHandler

from kube_copilot.chains import ReActLLM
from kube_copilot.prompts import get_analyze_prompt
from kube_copilot.labeler import CustomLLMThoughtLabeler

logging.basicConfig(stream=sys.stdout, level=logging.CRITICAL)
logging.getLogger().addHandler(logging.StreamHandler(stream=sys.stdout))


st.set_page_config(page_title="Analyze Kubernetes Manifests", page_icon="💬")
st.title("💬 Analyze Kubernetes Manifests")

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


namespace = st.text_input("Namespace", key="namespace",
                          placeholder="default", value="default")
resource_type = st.text_input(
    "Resource Type", key="resource_type", value="Pod")
resource_name = st.text_input(
    "Resource Name", key="resource_name", placeholder="nginx")

if st.button("Analyze"):
    if not os.getenv("OPENAI_API_KEY", ""):
        if not openai_api_key:
            st.info("Please add your OpenAI API key to continue.")
            st.stop()

        os.environ["OPENAI_API_KEY"] = openai_api_key
        os.environ["OPENAI_API_BASE"] = openai_api_base
        os.environ["GOOGLE_API_KEY"] = google_api_key
        os.environ["GOOGLE_CSE_ID"] = google_cse_id


    if not namespace or not resource_type or not resource_name:
        st.info(
            "Please add your namespace, resource_type and resource_name to continue.")
        st.stop()

    prompt = get_analyze_prompt(namespace, resource_type, resource_name)
    st_cb = StreamlitCallbackHandler(st.container(), thought_labeler=CustomLLMThoughtLabeler())
    chain = ReActLLM(model=model,
                     verbose=True,
                     enable_python=True,
                     auto_approve=True)

    response = chain.run(prompt, callbacks=[st_cb])
    st.markdown(response)
