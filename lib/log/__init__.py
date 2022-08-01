import datetime
import json as _json
import os
import sys

from lib import json_encoder

OPTIONS = {}


def set_option(key, value):
    '''set_option: set an optional globally
    '''
    OPTIONS[key] = value


class JSONEncoder(_json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, datetime.datetime):
            return obj.isoformat()

        return _json.JSONEncoder.default(self, obj)


def _msg(msg, **kwargs):
    '''_msg: format the message with the prefix and dry-run
    '''
    prefix = kwargs.pop('prefix', OPTIONS.get('prefix', True))
    dry_run = kwargs.pop('dry_run', OPTIONS.get('dry_run', False))

    delimiter = '='
    if 'delimiter' in kwargs:
        delimiter = kwargs.pop('delimiter')

    basename = os.path.basename(sys.argv[0])

    name = os.path.splitext(basename)[0]
    if dry_run:
        msg = 'DRY_RUN: {}'.format(msg)

    if prefix:
        msg = '{}: {}'.format(name, msg)

    for k, v in kwargs.items():
        msg += ' {}{}{}'.format(k, delimiter, v)

    return msg


def stdout(msg, **kwargs):
    '''stdout: write the message to stdout, adding a dry-run prefix and program prefix if applicable
    '''
    msg = _msg(msg, **kwargs)
    print(msg, file=sys.stdout, flush=True)


def stderr(msg, **kwargs):
    '''stderr: write the message to stderr, adding a dry-run prefix and program prefix if applicable
    '''
    msg = _msg(msg, **kwargs)

    print(msg, file=sys.stderr, flush=True)


def json(val, out='stdout', pretty=False, sort_keys=True):
    '''json: write the value to stdout or stderr as a json string
    '''
    fh = sys.stdout
    if out != 'stdout':
        fh = sys.stderr

    val = json_encoder.dumps(val)
    print(val, file=fh)
