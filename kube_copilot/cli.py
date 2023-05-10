#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import logging
import sys
import click
from kube_copilot.llm import init_openai
from kube_copilot.agent import CopilotLLM
from kube_copilot.prompts import (
    get_prompt,
    get_diagnose_prompt,
    get_audit_prompt
)


logging.basicConfig(stream=sys.stdout, level=logging.CRITICAL)
logging.getLogger().addHandler(logging.StreamHandler(stream=sys.stdout))


cmd_options = [
    click.option("--verbose", is_flag=True, default=False,
                 help="Enable verbose information of copilot execution steps"),
    click.option("--model", default="gpt-4",
                 help="OpenAI model to use for copilot execution, default is gpt-4"),
]


def add_options(options):
    '''Add options to a command'''
    def _add_options(func):
        for option in reversed(options):
            func = option(func)
        return func
    return _add_options


def get_llm_chain(verbose, model):
    '''Get Copilot LLM chain'''
    init_openai()
    return CopilotLLM(verbose=verbose, model=model)


@click.group()
@click.version_option()
def cli():
    '''Kubernetes Copilot powered by OpenAI'''


@cli.command(help="execute operations based on prompt instructions")
@click.argument('instructions')
@add_options(cmd_options)
def execute(instructions, verbose, model):
    '''Execute operations based on prompt instructions'''
    chain = get_llm_chain(verbose, model)
    result = chain.run(get_prompt(instructions))
    print(result)


@cli.command(help="diagnose problems for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@add_options(cmd_options)
def diagnose(namespace, pod, verbose, model):
    '''Diagnose problems for a Pod'''
    chain = get_llm_chain(verbose, model)
    result = chain.run(get_diagnose_prompt(namespace, pod))
    print(result)


@cli.command(help="audit security issues for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@add_options(cmd_options)
def audit(namespace, pod, verbose, model):
    '''Audit security issues for a Pod'''
    chain = get_llm_chain(verbose, model)
    result = chain.run(get_audit_prompt(namespace, pod))
    print(result)


def main():
    '''Main function'''
    cli()


if __name__ == "__main__":
    main()
