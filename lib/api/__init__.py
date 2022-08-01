import os

from lib import config
from lib import json_encoder

import requests
import urllib


BASE_URLS = {
    'dev': 'http://localhost:8665/internal-proxy-0',
    'prod': 'https://api.powertools.dev',
}


def build_url(uri):
    '''build_url: build a url and return the full version
    '''
    # NOTE: this doesn't work very well yet, and we only use it for email templates
    uri = uri.replace('/email-templates/', '')
    return 'https://email-templates.powertools.dev/{}'.format(uri)
    uri = uri = uri.replace('-templates/', '-templates-0/')
    return 'http://localhost:8665/{}'.format(uri)
    cfg = config.load()
    base_url = BASE_URLS.get(cfg.environment, BASE_URLS['dev'])
    base_url = base_url.rstrip('/')

    uri = uri.lstrip('/')

    return '{}/{}'.format(base_url, uri)


def do_request(uri, method='GET', data=None, json=False, ignore_err=False):
    '''do_request: perform a request
    '''
    url = build_url(uri)

    fn = {
        'get': requests.get,
        'head': requests.head,
        'post': requests.post,
        'put': requests.put,
        'delete': requests.delete,
    }[method.lower()]

    kwargs = {}
    if method in ('post', 'put') and data:
        if json:
            kwargs['data'] = json_encoder.dumps(data)
        else:
            kwargs['data'] = data

    # perform request
    resp = fn(url, **kwargs)
    is_err = resp.status_code / 4 == 100
    if is_err and not ignore_err:
        raise Exception('error requesting {}'.format(url))


    if not json:
        return resp.text

    return json_encoder.loads(resp.text)


def get_json(uri):
    '''get_json: perform a get request to fetch json
    '''
    return do_request(uri, method='GET', json=True)


def get(uri):
    '''get: get a url
    '''
    return do_request(uri, method='GET')
