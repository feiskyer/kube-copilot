# -*- coding: utf-8 -*-
import subprocess
from typing import List, Union
import tiktoken


class KubeProcess():
    '''Wrapper for cloud native commands.'''

    def __init__(self, command, max_tokens=3000, strip_newlines: bool = False, return_err_output: bool = False):
        """Initialize with stripping newlines."""
        self.strip_newlines = strip_newlines
        self.return_err_output = return_err_output
        self.command = command
        self.max_tokens = max_tokens
        self.encoding = tiktoken.encoding_for_model("gpt-3.5-turbo-0301")

    def run(self, args: Union[str, List[str]]) -> str:
        '''Run the command.'''
        if isinstance(args, str):
            args = [args]
        commands = ";".join(args)
        if not commands.startswith(self.command):
            commands = f'{self.command} {commands}'
        result = self.exec(commands)

        # TODO: workarounds for the following context length error with ChatGPT
        #   https://github.com/hwchase17/langchain/issues/2140
        #   https://github.com/hwchase17/langchain/issues/1767
        tokens = self.encoding.encode(result)
        while len(tokens) > self.max_tokens:
            result = result[:len(result) // 2]
            tokens = self.encoding.encode(result)
        return result

    def exec(self, commands: Union[str, List[str]]) -> str:
        """Run commands and return final output."""
        if isinstance(commands, str):
            commands = [commands]
        commands = ";".join(commands)
        try:
            output = subprocess.run(
                commands,
                shell=True,
                check=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
            ).stdout.decode()
        except subprocess.CalledProcessError as error:
            if self.return_err_output:
                return error.stdout.decode()
            return str(error)
        if self.strip_newlines:
            output = output.strip()
        return output
