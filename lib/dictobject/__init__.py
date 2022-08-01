import re

from lib import errors
from lib import log


def to_camelcase(obj):
    '''to_camelcase: convert an objects keys to camelCase
    '''
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


def to_underscore(obj):
    '''to_underscore: convert an objects keys to under_score
    '''
    if isinstance(obj, list):
        return [to_underscore(i) for i in obj]

    if not isinstance(obj, dict):
        return obj

    pattern = re.compile(r'([A-Z])')
    new_obj = {}

    for k, v in obj.items():
        k = pattern.sub(lambda x: '_' + x.group(1).lower(), k)
        new_obj[k] = v

        if isinstance(v, list) or isinstance(v, dict):
            new_obj[k] = to_underscore(v)

    return new_obj


def to_anyobject(obj, underscore=True):
    '''to_anyobject: accepts an object and returns a normalized anyobject with an optional converstion to camelcase
    '''
    if underscore:
        obj = to_underscore(obj)

    if isinstance(obj, list):
        return [to_anyobject(i) for i in obj]

    if not isinstance(obj, dict):
        return obj

    obj = AnyObject(obj)

    for k, v in list(obj.items()):
        if isinstance(v, dict) or isinstance(v, list):
            obj[k] = to_anyobject(v)

    return obj


class Error(errors.Error):
    status_code = 500

    def __init__(self, cls_name, attribute):
        self.cls_name = cls_name
        self.attribute = attribute

    def error(self):
        return 'attribute `{}` not found for class `{}`'.format(self.attribute, self.cls_name)


class DictObject(dict):
    '''DictObject: extends a dictionary object such that all values are accessible using object dot notation.

    This is useful to make objects both more friendly, but also easier to refactor later on. For instance, when a type
    becomes a more succinct class later on, we can easily swap it out by subclassing or redefining it's child class.
    Accessing the methods and attributes will remain the same.
    '''
    def __init__(self, *args, **kwargs):
        if len(args) and isinstance(args[0], dict):
            kwargs.update(args[0])

        required_keys = getattr(self.__class__, 'required_keys', [])

        for key in required_keys:
            if key not in kwargs:
                raise Error(self.__class__.__name__, key)

        super().__init__(**kwargs)

    def __getattr__(self, key):
        required_keys = getattr(self.__class__, 'required_keys', [])

        if key in required_keys and key not in self:
            raise Error(self.__class__.__name__, key)

        return self.get(key)

    def __setattr__(self, key, val):
        self[key] = val


class AnyObjectException(Exception):

    def __init__(self, key):
        message = '\nError: missing key `{}`'.format(key)
        super().__init__(message)


class AnyObject(dict):
    '''AnyObject: extends a dictionary and tries to access any key possible
    '''
    def __getattr__(self, key):
        if key in self:
            return self[key]

        try:
            super().getattr(key)
        except AttributeError as e:
            log.stderr('can not find key {}'.format(key))
            raise AnyObjectException(key)

    def __setattr__(self, key, val):
        self[key] = val
