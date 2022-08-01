import json
import os
import os.path
import pathlib
import time

import requests

from lib import log

# local configs for finding filepaths
BASE_DIR = pathlib.Path(__file__).parent.resolve()
QUERIES_DIR = os.path.join(BASE_DIR, 'queries')
MUTATIONS_DIR = os.path.join(BASE_DIR, 'mutations')

DEFAULT_ENV = 'dev'
BASE_URLS = {
    'stage': 'https://api.stage.nuon.co/query',
    'prod': 'https://api.nuon.co/query',
    'dev': 'http://localhost:8080/query',
}

API_TOKENS = {
    'stage': os.environ.get('APICTL_STAGE_TOKEN'),
    'prod': os.environ.get('APICTL_PROD_TOKEN'),
    'dev': os.environ.get('APICTL_DEV_TOKEN'),
}


def all_queries():
    '''all_queries: returns all of the known queries
    '''
    return [f for f in os.listdir(QUERIES_DIR)]


def all_queries_by_name():
    return [os.path.basename(f).split('.gql')[0] for f in all_queries()]


def all_mutations():
    '''all_mutations: returns all of the known queries
    '''
    return [f for f in os.listdir(MUTATIONS_DIR)]


def all_mutations_by_name():
    return [os.path.basename(f).split('.gql')[0] for f in all_mutations()]


def get_orig_filepath(typ, fn):
    '''get_orig_filepath: return an original filepath for the query or mutation
    '''
    if typ == 'query':
        return os.path.join(QUERIES_DIR, fn + '.gql')

    return os.path.join(MUTATIONS_DIR, fn + '.gql')


def get_orig_request_body(typ, fn, uservars={}, envvars=False):
    '''get_orig_request: return a reused request body
    '''
    fp = get_orig_filepath(typ, fn)
    with open(fp, 'r') as fh:
        contents = fh.read()

    if uservars:
        log.stderr('templating out user vars into request')
        for k, v in uservars.items():
            name = '${' + k.upper() + '}'
            contents = contents.replace(name, v)

    if envvars:
        log.stderr('templating out env vars into request')
        for k, v in os.environ.items():
            name = '${' + k.upper() + '}'
            contents = contents.replace(name, v)

    # run interactive prompt for values
    return contents


def get_request_body(typ, name, uservars={}, envvars=False):
    '''get_request_body: return a request body
    '''
    return get_orig_request_body(typ, name, uservars=uservars, envvars=envvars)


def do_request(env, typ, body):
    '''do_request: peform the request
    '''
    api_token = API_TOKENS[env]
    if not api_token:
        log.stderr('please set the envvar `APICTL_{}_TOKEN` with a valid token'.format(env.upper()))
        return

    query, variables = parse_request(body)
    url = BASE_URLS[env]

    log.stderr('using {}'.format(url))
    headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + api_token,
    }

    data = {'query': query, 'variables': variables}
    try:
        json_body = json.dumps(data)
    except Exception:
        log.json(data)
        logging.exception('unable to convert query to json')
        return

    try:
        resp = requests.post(url, data=json_body, headers=headers)
    except requests.exceptions.ConnectionError:
        log.stderr('unable to connect to {}'.format(url))

    return resp

def parse_request(body):
    '''parse_request: parse the request into a query + variables
    '''
    pieces = body.split('## variables')
    if len(pieces) < 2:
        return body, False, {}

    variables = {}
    for line in pieces[1].splitlines():
        if not line:
            continue

        k, v = line.split(':')
        variables[k] = v.strip()

    has_input_envelope = variables.pop('input_envelope', False)
    if has_input_envelope:
        variables = {'input': variables}

    return pieces[0], variables


def exec_request(typ, name, env, edit=False, uservars={}, envvars=True):
    '''exec: execute a query or mutation by name
    '''
    log.stderr('executing {} {} in {}'.format(typ, name, env))

    body = get_request_body(typ, name, envvars=envvars, uservars=uservars)
    if not body:
        return

    start_ts = time.time()
    resp = do_request(env, typ, body)
    try:
        resp.status_code
    except:
        log.stderr('unable to determine status code of response')
        return

    e2e_latency = time.time() - start_ts
    render_response(e2e_latency, resp, headers=True)


def render_errors(errors):
    '''render_errors: render errors
    '''
    log.stderr('error response')
    for error in errors:
        if 'message' in error and 'backtrace' in error:
            log.stderr('error: {}'.format(error['message']))
            log.stderr('backtrace:')
            for line in reversed(error['backtrace']):
                log.stderr(line)
        else:
            log.json(error)


def render_response(e2e_latency, resp, headers=True):
    '''render_response: render api response to stderr / stdout
    '''
    log.stderr('response status code {}'.format(resp.status_code))
    log.stderr('e2e latency {:.4f}ms'.format(e2e_latency * 1000))

    if headers:
        for k, v in resp.headers.items():
            log.stderr('header {} => {}'.format(k, v))

    try:
        json_data = resp.json()
    except Exception:
        content = resp.content.decode('utf-8')
        json_data = json.loads(content)

    if json_data.get('errors'):
        render_errors(json_data['errors'])
    else:
        log.json(json_data)
