#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import click
from kube_copilot.llm import init_openai
from kube_copilot.agent import CopilotLLM
from kube_copilot.prompts import (
    get_prompt,
    get_diagnose_prompt,
    get_audit_prompt
)


cmd_options = [
    click.option("--short", is_flag=True, default=False,
                 help="Disable verbose information of copilot execution steps"),
    click.option("--model", default="gpt-3.5-turbo",
                 help="OpenAI model to use for copilot execution, default is gpt-3.5-turbo"),
    click.option("--enable-terminal", is_flag=True, default=False,
                 help="Enable Copilot to run programs within terminal. Enable with caution since Copilot may execute inappropriate commands"),
]


def add_options(options):
    '''Add options to a command'''
    def _add_options(func):
        for option in reversed(options):
            func = option(func)
        return func
    return _add_options


def get_llm_chain(verbose, model, enable_terminal):
    '''Get Copilot LLM chain'''
    init_openai()
    return CopilotLLM(verbose=verbose, model=model, enable_terminal=enable_terminal)


@click.group()
@click.version_option()
def cli():
    '''Kubernetes Copilot powered by OpenAI'''


@cli.command(help="execute operations based on prompt instructions")
@click.argument('instructions')
@add_options(cmd_options)
def execute(instructions, short, model, enable_terminal):
    '''Execute operations based on prompt instructions'''
    if click.confirm("Copilot may generate and execute inappropriate operations steps, are you sure to continue?"):
        chain = get_llm_chain(not short, model, enable_terminal)
        result = chain.run(get_prompt(instructions))
        print(result)


@cli.command(help="diagnose problems for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@add_options(cmd_options)
def diagnose(namespace, pod, short, model, enable_terminal):
    '''Diagnose problems for a Pod'''
    chain = get_llm_chain(not short, model, enable_terminal)
    result = chain.run(get_diagnose_prompt(namespace, pod))
    print(result)


@cli.command(help="audit security issues for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@add_options(cmd_options)
def audit(namespace, pod, short, model, enable_terminal):
    '''Audit security issues for a Pod'''
    chain = get_llm_chain(not short, model, enable_terminal)
    result = chain.run(get_audit_prompt(namespace, pod))
    print(result)


def main():
    '''Main function'''
    cli()


if __name__ == "__main__":
    main()
