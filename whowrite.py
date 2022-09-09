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
from concurrent.futures import ThreadPoolExecutor
import click


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


def handle(filepath: str, a):
    cmd = f"git blame {filepath}"
    try:
        resp = sp.check_output(cmd, shell=True)
        resp = str(resp, encoding='utf8')
    except Exception as ex:
        print(ex, f'run command {cmd} error')
        return
    lines = resp.splitlines()
    for line in lines:
        fields = re.search('\((.+?)\)', line).group(1).split()
        name = fields[0]
        a[name] += 1


def main(worker=15):
    begin_time = time.time()
    a = defaultdict(lambda: 0)
    with ThreadPoolExecutor(worker) as pool:
        for parent, folders, files in walk('.', should_enter):
            for f in files:
                _, ext = splitext(basename(f))
                if ext in ('.py', '.go', '.js', '.cs', '.cpp', '.c', '.cxx',):
                    filepath = join(parent, f)
                    pool.submit(handle, filepath, a)
        pool.shutdown(True)
    pprint(a)
    end_time = time.time()
    print('Time used', end_time - begin_time)


if __name__ == '__main__':
    main()
