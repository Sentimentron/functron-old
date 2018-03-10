from setuptools import setup, find_packages
from codecs import open
from os import path

here = path.abspath(path.dirname(__file__))

with open(path.join(here, 'README.rst'), encoding='utf-8') as f:
    long_description = f.read()

setup(
    name='pyfunctron',
    version='0.1.0',
    description='Python bindings for Functron - a minimalist FaaS',
    long_description=long_description,
    url='https://github.com/Sentimentron/functron/python',
    author_email='richard@sentimentron.co.uk',
    classifiers=[
        'Development Status :: 3 - Alpha',
        'Intended Audience :: Developers',
        'Topic :: Software Development',
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python :: 3.4'
   ],
   keywords='docker faas deployment microservices',
   py_modules=['pyfunctron'],
   install_requires=['requests'],
   project_urls={
    'Bug Reports': 'https://github.com/Sentimentron/functron/issues',
    'Source': 'https://github.com/Sentimentron/functron'
   },
)
