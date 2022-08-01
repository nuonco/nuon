import json

from lib import config
from lib import log

import requests


def default_filter(data):
    '''default_filter: return whether a message should be sent
    '''
    cfg = config.load()

    if 'email' in data and data.get('email') in cfg.slack_ignore_users:
        return

    return True


def send_message(msg_tmpl, data=None, filter_fn=default_filter, webhook_key='slack_webhook_url'):
    '''send_message: send a slack message
    '''
    if filter_fn and data:
        if not filter_fn(data):
            log.stderr('skipping send per filter function')
            return

    cfg = config.load()

    msg = msg_tmpl
    if data:
        try:
            msg = msg_tmpl.format(**data)
        except Exception:
            if isinstance(data, list):
                msg = msg_tmpl.format(*data)
            else:
                keys = list(data.values())
                msg = msg_tmpl.format(*keys)

    data = {
        'text': msg,
    }
    headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
    }

    webhook_url = getattr(cfg, webhook_key)
    if not webhook_url:
        log.stderr('no webhook url found in config for {}'.format(webhook_key))
        return

    response = requests.post(webhook_url, data=json.dumps(data), headers=headers)
    if response.status_code != 200:
        print('slack response error code {}'.format(response.status_code))


send_msg = send_message
