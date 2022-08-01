import collections
import datetime
import json
import re


class JSONEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, bytes):
            return obj.decode('utf-8')

        if isinstance(obj, datetime.datetime):
            return int(obj.timestamp())

        if isinstance(obj, collections.abc.KeysView):
            obj = [str(key) for key in obj]
            return list(obj)

        return json.JSONEncoder.default(self, obj)


def to_camelcase(obj):
    if isinstance(obj, list):
        return [to_camelcase(i) for i in obj]

    if not isinstance(obj, dict):
        return obj

    pattern = re.compile(r'_([a-z])')
    new_obj = {}

    for k, v in obj.items():
        k = pattern.sub(lambda x: x.group(1).upper(), k)
        new_obj[k] = v

        if isinstance(v, list) or isinstance(v, dict):
            new_obj[k] = to_camelcase(v)

    return new_obj


def loads(content, camelcase=False):
    '''loads: load a string as json
    '''
    obj = json.loads(content)

    if camelcase:
        obj = to_camelcase(obj)

    return obj


def dumps(obj, camelcase=False, pretty=True):
    '''dumps: dump an object to json
    '''
    if camelcase:
        obj = to_camelcase(obj)

    kwargs = {}
    if pretty:
        kwargs['sort_keys'] = True
        kwargs['indent'] = 4
        kwargs['separators'] = (',', ': ')

    return json.dumps(obj, cls=JSONEncoder, **kwargs)
