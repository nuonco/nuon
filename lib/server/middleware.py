import datetime
import functools
import logging
import os
import re

import flask

from lib.server import errors as errors


def token_auth(fn):
    '''auth_handler: raises an error if a valid token is not provided. Reads
    tokens from the environment variable `TOKENS`
    '''
    @functools.wraps(fn)
    def _(*args, **kwargs):
        token = flask.request.headers.get('x-auth-token')
        if not token:
            raise errors.TokenAuthError()

        # NOTE: safe exec encodes lists from configs as `[TOKEN TOKEN TOKEN]`,
        # until this changes, the following splitting is required.
        valid_tokens = os.environ.get('TOKENS')
        if not valid_tokens:
            raise errors.TokenConfigError()

        valid_tokens = valid_tokens.split(',')

        if token not in valid_tokens:
            raise errors.TokenAuthError()

        return fn(*args, **kwargs)

    return _


def error(fn):
    '''error_handler: catches exceptions, rendering them into well formed
    errors
    '''
    @functools.wraps(fn)
    def _(*args, **kwargs):
        try:
            res = fn(*args, **kwargs)
            return res
        except errors.Error as e:
            status_code = e.status_code
            res = {
                'error': e.error(),
            }
        except Exception as e:
            logging.exception(e)
            status_code = 500
            res = {
                'error': str(e)
            }

        return flask.jsonify(res), status_code

    return _


def _to_camelcase(obj):
    if isinstance(obj, list):
        return [_to_camelcase(i) for i in obj]

    if not isinstance(obj, dict):
        return obj

    pattern = re.compile(r'_([a-z])')
    new_obj = {}

    for k, v in obj.items():
        k = pattern.sub(lambda x: x.group(1).upper(), k)
        new_obj[k] = v

        if isinstance(v, list) or isinstance(v, dict):
            new_obj[k] = _to_camelcase(v)

    return new_obj


def _to_underscore(obj):
    if isinstance(obj, list):
        return [_to_underscore(i) for i in obj]

    if not isinstance(obj, dict):
        return obj

    pattern = re.compile(r'([A-Z])')
    new_obj = {}

    for k, v in obj.items():
        k = pattern.sub(lambda x: '_' + x.group(1).lower(), k)
        new_obj[k] = v

        if isinstance(v, list) or isinstance(v, dict):
            new_obj[k] = _to_underscore(v)

    return new_obj


def _timestamps_to_epochs(obj):
    if isinstance(obj, list):
        return [_timestamps_to_epochs(i) for i in obj]

    if not isinstance(obj, dict):
        return obj

    for k, v in obj.items():
        if isinstance(v, datetime.datetime):
            obj[k] = int(v.timestamp() * 1000)

        if isinstance(v, list) or isinstance(v, dict):
            obj[k] = _timestamps_to_epochs(v)

    return obj


def json(fn):
    '''json: ensures that the content-type is json and calls
    flask.jsonify for all return values
    '''
    @functools.wraps(fn)
    def _(*args, **kwargs):
        accept = str(flask.request.accept_mimetypes)
        content_type = flask.request.content_type

        if not accept.lower().startswith('application/json'):
            raise errors.InvalidAcceptMimeTypesError(accept)

        json_content_type = content_type and content_type.lower().startswith('application/json')
        if flask.request.method not in ('GET', 'DELETE', 'HEAD') and not json_content_type:
            raise errors.InvalidContentTypeError(content_type)

        json_fmt = flask.request.headers.get('x-json-fmt', 'snake')
        if json_fmt.lower() == 'camelcase' and flask.request.method in ('PUT', 'PATCH', 'POST'):
            data = flask.request.get_json()
            kwargs['data'] = _to_underscore(data)

        res = fn(*args, **kwargs)

        if json_fmt.lower() == "camelcase":
            res = _to_camelcase(res)

        ts_fmt = flask.request.headers.get('x-timestamp-fmt', 'utc')
        if ts_fmt == "epoch":
            res = _timestamps_to_epochs(res)

        return flask.jsonify(res)

    return _
