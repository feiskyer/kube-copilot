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
@click.pass_context
def cli(ctx, short):
    init_openai()
    ctx.ensure_object(dict)
    ctx.obj["chain"] = CopilotLLM(verbose = not short)


@cli.command(help="execute operations based on prompt instructions")
@click.argument('instructions')
@click.pass_context
def execute(ctx, instructions):
    if click.confirm("Copilot may generate and execute inappropriate operations steps, are you sure to continue?"):
        chain = ctx.obj["chain"]
        result = chain.run(get_prompt(instructions))
        print(result)

@cli.command(help="diagnose problems for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@click.pass_context
def diagnose(ctx, namespace, pod):
    chain = ctx.obj["chain"]
    result = chain.run(get_diagnose_prompt(namespace, pod))
    print(result)

@cli.command(help="audit security issues for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@click.pass_context
def audit(ctx, namespace, pod):
    chain = ctx.obj["chain"]
    result = chain.run(get_audit_prompt(namespace, pod))
    print(result)


def main():
    cli()


if __name__ == "__main__":
    main()
