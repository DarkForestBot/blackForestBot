#!/usr/bin/env python3
from os.path import join
import base64, os, sys

BUILD_PATH = os.environ.get('BUILD_PATH')
if not BUILD_PATH:
    print("ERR: No BUILD_PATH specified.")
    sys.exit(1)

CODE = ''

for root, dir, file in os.walk(join(BUILD_PATH, 'assets', 'images')):
    for image in file:
        file_name = image.split('.')[0].capitalize()
        with open(join(root, image), 'rb') as f:
            data = base64.b64encode(f.read())
            CODE += 'var Image{0} = "{1}"\n'.format(file_name, data.decode('utf-8'))

with open('{0}/basis/images.go.in'.format(BUILD_PATH), 'r') as f:
    s = f.read()
    s = s.replace('%SOMECODE%', CODE)
    with open('{0}/basis/images.go'.format(BUILD_PATH), 'w') as w:
        w.write(s)