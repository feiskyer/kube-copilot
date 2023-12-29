#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import logging
import sys
import click
from kube_copilot.chains import ReActLLM
from kube_copilot.shell import KubeProcess
from kube_copilot.prompts import (
    get_prompt,
    get_diagnose_prompt,
    get_analyze_prompt,
    get_audit_prompt,
    get_generate_prompt
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


@click.group()
@click.version_option()
def cli():
    '''Kubernetes Copilot powered by OpenAI'''


@cli.command(help="execute operations based on prompt instructions")
@click.argument('instructions')
@add_options(cmd_options)
def execute(instructions, verbose, model):
    '''Execute operations based on prompt instructions'''
    chain = ReActLLM(verbose=verbose, model=model, enable_python=True)
    result = chain.run(get_prompt(instructions))
    print(result)


@cli.command(help="diagnose problems for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@add_options(cmd_options)
def diagnose(namespace, pod, verbose, model):
    '''Diagnose problems for a Pod'''
    # chain = PlanAndExecuteLLM(verbose=verbose, model=model, enable_python=True)
    chain = ReActLLM(verbose=verbose, model=model, enable_python=True)
    result = chain.run(get_diagnose_prompt(namespace, pod))
    print(result)


@cli.command(help="audit security issues for a Pod")
@click.argument('pod')
@click.argument('namespace', default="default")
@add_options(cmd_options)
def audit(namespace, pod, verbose, model):
    '''Audit security issues for a Pod'''
    chain = ReActLLM(verbose=verbose, model=model)
    result = chain.run(get_audit_prompt(namespace, pod))
    print(result)


@cli.command(help="analyze issues for a given resource")
@click.argument('resource')
@click.argument('name')
@click.argument('namespace', default="default")
@add_options(cmd_options)
def analyze(resource, namespace, name, verbose, model):
    '''Analyze potential issues for a given resource'''
    chain = ReActLLM(verbose=verbose, model=model)
    result = chain.run(get_analyze_prompt(namespace, resource, name))
    print(result)


@cli.command(help="generate Kubernetes manifests")
@click.argument('instructions')
@add_options(cmd_options)
def generate(instructions, verbose, model):
    '''Generate Kubernetes manifests'''
    chain = ReActLLM(verbose=verbose, model=model)
    result = chain.run(get_generate_prompt(instructions))
    print(result)

    # Apply the generated manifests in cluster
    if click.confirm('Do you approve to apply the generated manifests to cluster?'):
        manifests = result.removeprefix(
            '```').removeprefix('yaml').removesuffix('```')
        print(KubeProcess(command="kubectl").run(
            'kubectl apply -f -', input=bytes(manifests, 'utf-8')))


def main():
    '''Main function'''
    cli()


if __name__ == "__main__":
    main()
