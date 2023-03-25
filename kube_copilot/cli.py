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


@click.group()
@click.version_option()
@click.option("--short", is_flag=True, default=False, help="Disable verbose information of copilot execution steps")
@click.option("--model", default="gpt-3.5-turbo", help="OpenAI model to use for copilot execution, default is gpt-3.5-turbo")
@click.option("--enable-terminal", is_flag=True, default=False, help="Enable Copilot to run programs within terminal. Enable with caution since Copilot may execute inappropriate commands")
@click.pass_context
def cli(ctx, short, model, enable_terminal):
    '''Kubernetes Copilot powered by OpenAI'''
    init_openai()
    ctx.ensure_object(dict)
    ctx.obj["chain"] = CopilotLLM(
        verbose=not short, model=model, enable_terminal=enable_terminal)


@cli.command(help="execute operations based on prompt instructions")
@click.argument('instructions')
@click.pass_context
def execute(ctx, instructions):
    '''Execute operations based on prompt instructions'''
    if click.confirm("Copilot may generate and execute inappropriate operations steps, are you sure to continue?"):
        chain = ctx.obj["chain"]
        result = chain.run(get_prompt(instructions))
        print(result)


@cli.command(help="diagnose problems for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@click.pass_context
def diagnose(ctx, namespace, pod):
    '''Diagnose problems for a Pod'''
    chain = ctx.obj["chain"]
    result = chain.run(get_diagnose_prompt(namespace, pod))
    print(result)


@cli.command(help="audit security issues for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@click.pass_context
def audit(ctx, namespace, pod):
    '''Audit security issues for a Pod'''
    chain = ctx.obj["chain"]
    result = chain.run(get_audit_prompt(namespace, pod))
    print(result)


def main():
    cli()


if __name__ == "__main__":
    main()
