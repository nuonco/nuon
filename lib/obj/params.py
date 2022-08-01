import datetime

import pytz


class FieldMissingErr(Exception):
    def __init__(self, field):
        self.field = field

    def __str__(self):
        return 'field `{}` is missing'.format(self.field)


class FieldEmptyErr(Exception):
    def __init__(self, field):
        self.field = field

    def __str__(self):
        return 'field `{}` is empty'.format(self.field)


class FieldDateErr(Exception):
    def __init__(self, field, value):
        self.field = field
        self.value = value

    def __str__(self):
        return 'field `{}` is not a valid date - `{}`'.format(self.field, self.value)


def ensure_int(data, fields):
    '''ensure_int: ensure each field is a valid integer
    '''
    for field in fields:
        if field not in data:
            raise FieldMissingErr(field)

        val = int(data[field])
        data[field] = val

    return data


def ensure_optional(data, fields):
    '''ensure_optional: ensure that each field is set on the dictionary, if it's optional and missing it will be set to
    none
    '''
    for field in fields:
        if field not in data:
            data[field] = None

        if not data[field]:
            data[field] = None


def ensure(data, fields):
    '''ensure: by default, we have a convention of no more than three parameters to a single shared/* method, for
    readability. This method takes a params dict and parses the required fields, throwing an error if they don't exist
    '''
    for field in fields:
        if field not in data:
            raise FieldMissingErr(field)

        if not data[field]:
            raise FieldEmptyErr(field)


def ensure_boolean(data, fields):
    '''ensure_boolean: ensure boolean fields, if the field does not exist set it's value as false
    '''
    for field in fields:
        if field not in data:
            data[field] = False
            continue

        if data[field] == 'off':
            data[field] = False
        elif data[field] == 'on':
            data[field] = True

    return data


def ensure_date(data, fields):
    '''ensure_date: ensure the given fields are date fields
    '''
    ensure(data, fields)

    fmts = ['%b %d, %Y']
    for field in fields:
        if isinstance(data[field], datetime.datetime):
            continue

        for fmt in fmts:
            try:
                dt = datetime.datetime.strptime(data[field], fmt)
                tz = pytz.timezone('America/Los_Angeles')
                dt = tz.localize(dt)
                data[field] = dt
                break
            except ValueError:
                continue
        else:
            raise FieldDateErr(field, data[field])

    return data
