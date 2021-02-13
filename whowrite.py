"""
一个repo的代码是谁写的
利用gitblame命令
"""
import os
import time
from pprint import pprint
import subprocess as sp
from collections import Counter, defaultdict
from os.path import *
from typing import Callable
import re


def walk(path, should_enter: Callable):
    if isfile(path):
        raise Exception(f"{path} is a file")
    folders = []
    files = []
    for son in os.listdir(path):
        if isfile(join(path, son)):
            files.append(son)
        else:
            folders.append(son)
            if should_enter(join(path, son)):
                for i in walk(join(path, son), should_enter):
                    yield i
    yield path, folders, files


def should_enter(folder):
    if isfile(folder):
        return False
    if basename(folder).startswith('.'):
        return False
    if basename(folder) == 'node_modules':
        return False
    return True


def main():
    a = defaultdict(lambda: 0)
    for parent, folders, files in walk('.', should_enter):
        for f in files:
            if f.endswith('.py') or f.endswith('.go') or f.endswith('.js'):
                filepath = join(parent, f)
                cmd = f"git blame {filepath}"
                try:
                    resp = sp.check_output(cmd, shell=True)
                    resp = str(resp, encoding='utf8')
                except Exception as ex:
                    print(ex, f'run command {cmd} error')
                    continue
                lines = resp.splitlines()
                for line in lines:
                    fields = re.search('\((.+?)\)', line).group(1).split()
                    name = fields[0]
                    a[name] += 1
    pprint(a)


begin_time = time.time()
main()
end_time = time.time()
print('Time used', end_time - begin_time)
