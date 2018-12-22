#!/usr/bin/env python3
# -*- coding:utf-8 -*-

import re
from os import listdir
from os.path import isfile, join
import json

file_dir = "./post/"
index_file = "./index.json"


def process(fname):
    meta_info = dict()
    meta_info['file_name'] = fname
    with open(file_dir+fname) as f:
        content = f.read()
        result = re.match(r'<!-+\n(.*)\n-+>\n', content, re.S)
        if result:
            metas = result.groups()[0].split('\n')
            for info in metas:
                [key, value] = info.split('::')
                if key == 'title'or key == 'date':
                    meta_info[key] = value.strip()
                elif key == 'categories' or key == 'tags':
                    meta_info[key] = list(map(lambda x: x.strip(), value.split(',')))
                    print('here', meta_info[key])
        else:
            print('meta info wrong, file name', fname)
            exit()
    print(meta_info)
    return meta_info


files = [f for f in listdir(file_dir) if isfile(join(file_dir, f))]
md_files = filter(lambda f : f.endswith('.md'), files)
metas = []
for md in md_files:
    meta_info = process(md)
    metas.append(meta_info)
metas = sorted(metas, key = lambda k: k['date'])
metas.reverse()

for meta in metas:
    meta['date'] = re.findall(r'\d{4}-\d{2}-\d{2}', meta['date'])[0]

with open(index_file, 'r+') as f:
    data = json.load(f)
    data['posts'] = metas
    f.seek(0)
    f.truncate()
    json.dump(data, f, indent=2)

print(metas)