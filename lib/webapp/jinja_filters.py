import calendar
import datetime
import json

from lib import json_encoder

import pytz


def count_filter(objs, key, value):
    '''count_filter: return the count of objects in the list that meet the key/value requirement
    '''
    filtered = [o for o in objs if o.get(key) == value]
    return len(filtered)


def sort_by(objs, key, desc=True):
    return sorted(objs, key=lambda o: o[key], reverse=desc)


def to_pretty_json(value):
    return json.dumps(value, sort_keys=True,
                      indent=4, separators=(',', ': '))


def to_json(value):
    return json_encoder.dumps(value)


def datetime_to_timestamp(d):
    timezone = pytz.timezone('America/Los_Angeles')
    dt = d.astimezone(timezone)
    return dt.strftime('%Y-%m-%d %-I:%M%p')


def datetime_to_human_delta(d):
    '''datetime_to_human_delta: convert the date time to unit ago
    '''
    def plural(val):
        if int(val) == 1:
            return ''
        return 's'

    now = datetime.datetime.now(pytz.UTC)
    delta = now - d

    if delta.days:
        return '{} day{} ago'.format(delta.days, plural(delta.days))

    val = delta.seconds / 3600
    if delta.seconds / 3600 >= 1:
        return '{} hour{} ago'.format(int(val), plural(val))

    val = delta.seconds / 60
    if val >= 1:
        return '{} minute{} ago'.format(int(val), plural(val))

    return '{} second{} ago'.format(int(delta.seconds), plural(delta.seconds))


def datetime_to_date_string(d):
    '''datetime_to_date_string: create a human friendly date string from the data
    '''
    weekday = calendar.day_abbr[d.weekday()]
    month = calendar.month_name[d.month]

    return '{}, {} {}, {}'.format(weekday, month, d.day, d.year)


def datetime_to_uri(d):
    '''datetime_to_uri: return a datetime uri that can be parsed by this tool
    '''
    return '{}/{}/{}'.format(d.year, d.month, d.day)


def datetime_to_datepicker_string(dt):
    fmts = ['%Y_%m_%d', '%Y/%m/%d']
    if isinstance(dt, str):
        for fmt in fmts:
            try:
                dt = datetime.datetime.strptime(dt, fmt)
                return datetime_to_datepicker_string(dt)
            except ValueError:
                continue

        return dt

    return dt.strftime('%b %d, %Y')

def init(app):
    '''init: register all jinja filters
    '''
    app.jinja_env.filters['to_pretty_json'] = to_pretty_json
    app.jinja_env.filters['to_json'] = to_json
    app.jinja_env.filters['datetime_to_timestamp'] = datetime_to_timestamp
    app.jinja_env.filters['datetime_to_human_delta'] = datetime_to_human_delta
    app.jinja_env.filters['datetime_to_date_string'] = datetime_to_date_string
    app.jinja_env.filters['datetime_to_uri'] = datetime_to_uri
    app.jinja_env.filters['datetime_to_datepicker_string'] = datetime_to_datepicker_string
    app.jinja_env.filters['count_filter'] = count_filter
    app.jinja_env.filters['sort_by'] = sort_by

