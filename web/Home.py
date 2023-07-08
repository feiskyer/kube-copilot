# -*- coding: utf-8 -*-
import logging
import sys
import os

import streamlit as st
from langchain.callbacks import StreamlitCallbackHandler

from kube_copilot.chains import ReActLLM
from kube_copilot.llm import init_openai
from kube_copilot.prompts import get_prompt
from kube_copilot.kubeconfig import setup_kubeconfig
from kube_copilot.labeler import CustomLLMThoughtLabeler

# setup logging
logging.basicConfig(stream=sys.stdout, level=logging.WARNING)
logging.getLogger().addHandler(logging.StreamHandler(stream=sys.stdout))

# setup kubeconfig when running inside Pod
setup_kubeconfig()

st.set_page_config(page_title="Kubernetes Copilot", page_icon="💬")
st.title("💬 Kubernetes Copilot")

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


if "messages" not in st.session_state:
    st.session_state["messages"] = [
        {"role": "assistant", "content": "I'm your Kubernetes Copilot. How can I help you?"}]

for msg in st.session_state.messages:
    st.chat_message(msg["role"]).write(msg["content"])

if prompt := st.chat_input():
    if not os.getenv("OPENAI_API_KEY", ""):
        if not openai_api_key:
            st.info("Please add your OpenAI API key to continue.")
            st.stop()

        os.environ["OPENAI_API_KEY"] = openai_api_key
        os.environ["OPENAI_API_BASE"] = openai_api_base
        os.environ["GOOGLE_API_KEY"] = google_api_key
        os.environ["GOOGLE_CSE_ID"] = google_cse_id

    init_openai()

    st.session_state.messages.append({"role": "user", "content": prompt})
    st.chat_message("user").write(prompt)

    st_cb = StreamlitCallbackHandler(
        st.container(), thought_labeler=CustomLLMThoughtLabeler())
    chain = ReActLLM(model=model,
                     verbose=True,
                     enable_python=True,
                     auto_approve=True)

    with st.chat_message("assistant"):
        response = chain.run(get_prompt(prompt), callbacks=[st_cb])
        st.session_state.messages.append(
            {"role": "assistant", "content": response})
        st.write(response)
