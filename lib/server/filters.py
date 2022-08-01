import calendar
import datetime
import time
import json

import yaml
import pytz


def day_delta_to_unix_timestamp(days):
    '''hour_delta_to_unix_epoch: return a unix epoch for the delta
    '''
    return hour_delta_to_unix_timestamp(days * 24)


def hour_delta_to_unix_timestamp(hours):
    '''hour_delta_to_unix_epoch: return a unix epoch for the delta
    '''
    ts = int(time.time())

    return ts - (hours * 3600)


def count_filter(objs, key, value):
    '''count_filter: return the count of objects in the list that meet the key/value requirement
    '''
    filtered = [o for o in objs if o.get(key) == value]
    return len(filtered)


def to_pretty_json(value):
    return json.dumps(value, sort_keys=True,
                      indent=4, separators=(',', ': '))


def datetime_to_timestamp(d):
    timezone = pytz.timezone('America/Los_Angeles')
    dt = d.astimezone(timezone)
    return dt.strftime('%Y-%m-%d %-I:%M%p')


def plural(val):
    if int(val) == 1:
        return ''

    return 's'


def str_or_dict_field(val, key='raw'):
    if not val:
        return ''

    if isinstance(val, str):
        return val

    return val.get(key, '')


def datetime_to_human_delta(d):
    '''datetime_to_human_delta: convert the date time to unit ago
    '''

    now = datetime.datetime.now(pytz.UTC)
    delta = now - d

    human_delta = to_human_delta(delta)
    return '{} ago'.format(human_delta)


def to_human_delta(delta):
    if not delta:
        return None

    if delta.days:
        return '{} day{}'.format(delta.days, plural(delta.days))

    val = delta.seconds / 3600
    if delta.seconds / 3600 >= 1:
        return '{} hour{}'.format(int(val), plural(val))

    val = delta.seconds / 60
    if val >= 1:
        return '{} minute{}'.format(int(val), plural(val))

    return '{} second{}'.format(int(delta.seconds), plural(delta.seconds))


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


def to_url(obj):
    '''to_url: convert an object to a url
    '''
    if not obj:
        return ''

    if obj.startswith('http'):
        return obj

    return 'https://' + obj


def to_yaml(obj):
    '''to_yaml: convert an object to a yaml string
    '''
    if not obj:
        return ''

    obj = dict(obj)
    return yaml.dump(obj, default_flow_style=False, default_style='|')


def to_key_value_str(string, delimiter='__'):
    '''to_key_value_str: turn a string into a `key = value` from a delimiter
    '''
    pieces = string.split(delimiter)
    return '{} = {}'.format(pieces[0], pieces[1])


def auto_detect_link(string, target='_blank'):
    '''auto_detect_link: auto detect a link
    '''
    if not string.startswith('http') and not string.startswith('www'):
        return string

    return '<a target={} href="{}">{}</a>'.format(target, string, string)


def init_jinja_filters(app):
    '''init_jinja_filters: register all jinja filters
    '''
    app.jinja_env.filters['tojson_pretty'] = to_pretty_json
    app.jinja_env.filters['to_json'] = to_pretty_json
    app.jinja_env.filters['datetime_to_timestamp'] = datetime_to_timestamp
    app.jinja_env.filters['datetime_to_human_delta'] = datetime_to_human_delta
    app.jinja_env.filters['datetime_to_date_string'] = datetime_to_date_string
    app.jinja_env.filters['datetime_to_uri'] = datetime_to_uri
    app.jinja_env.filters['datetime_to_datepicker_string'] = datetime_to_datepicker_string
    app.jinja_env.filters['count_filter'] = count_filter
    app.jinja_env.filters['to_yaml'] = to_yaml
    app.jinja_env.filters['to_human_delta'] = to_human_delta
    app.jinja_env.filters['str_or_dict_field'] = str_or_dict_field
    app.jinja_env.filters['to_url'] = to_url
    app.jinja_env.filters['to_key_value_str'] = to_key_value_str
    app.jinja_env.filters['auto_detect_link'] = auto_detect_link
    app.jinja_env.filters['hour_delta_to_unix_timestamp'] = hour_delta_to_unix_timestamp
    app.jinja_env.filters['day_delta_to_unix_timestamp'] = day_delta_to_unix_timestamp
