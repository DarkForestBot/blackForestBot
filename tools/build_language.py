#!/usr/bin/env python3
from os.path import join
import os, sys

BUILD_PATH = os.environ.get('BUILD_PATH')
if not BUILD_PATH:
    print("ERR: No BUILD_PATH specified.")
    sys.exit(1)

CODE = ''

for root, dir, file in os.walk(join(BUILD_PATH, 'assets', 'language')):
    for lang in file:
        file_name = lang.split('.')[0].replace('-', '').capitalize()
        VAR = "LocaleLanguage{0}".format(file_name)
        CODE += '{0} := Language'.format(VAR); CODE += '{\n'
        LANGNAME = ''

        for i in open(join(root, lang), 'r'):
            i = i.replace('\r', '').replace('\n', '').replace('\x08', '')
            if i.strip() == "":
                continue
            if i[0] == "#":  # comment
                continue

            op = i.split('=', 1)[0].strip()
            val = i.split('=', 1)[1]
            
            if op.split(':', 1)[0] == 'meta':
                if op.split(':', 1)[1] == 'language':
                    LANGNAME = val
            elif op.split(':', 1)[0] == 'trans':
                CODE += '        "{0}": "{1}",\n'.format(op.split(':', 1)[1], val)
        CODE += '    }\n'
        CODE += '    GlobalLanguageList["{0}"] = {1}\n'.format(LANGNAME, VAR)
        CODE += '    '

print(CODE)

with open('{0}/basis/language.go.in'.format(BUILD_PATH), 'r') as f:
    s = f.read()
    s = s.replace('%SOMECODE%', CODE)
    with open('{0}/basis/language.go'.format(BUILD_PATH), 'w') as w:
        w.write(s)
