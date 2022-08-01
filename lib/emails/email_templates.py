import json
import uuid
import re

from lib import api
from lib import log


def fetch_all_templates():
    '''fetch_all_templates: fetch all templates
    '''
    try:
        resp = api.get_json('/email-templates')
    except Exception:
        return []

    return resp['templates']


def fmt(tmpl_id, tmpl):
    '''fmt: format a template
    '''
    tmpl_vars = set([])
    matches = re.findall('\[.*?\]', tmpl)
    for match in matches:
        match = match.replace('[', '').replace(']', '')
        if not match.startswith('var_'):
            continue

        tmpl_vars.add(match)

    return {
        'html': tmpl,
        'subject': tmpl_id,
        'vars': list(tmpl_vars),
        'id': tmpl_id,
    }


def render_template(tmpl_id, params):
    '''render_template: render the variables into the message and return it as a string
    '''
    tmpl = fetch_template(tmpl_id)
    msg = tmpl['html']

    for var in tmpl['vars']:
        key = var.replace('var_', '')
        if not params.get(key):
            raise Exception('Missing var {}'.format(var))

        msg = msg.replace('[{}]'.format(var), params[key])

    return msg


def new_email_from_template(tmpl_id, params={}, user_email=None):
    '''new_email_from_template: create a new email from a template id
    '''
    tmpl = fetch_template(tmpl_id)
    sender_id = 'jon_powertoolsdev'

    email = {
        'status': 'new',
        'email_id': str(uuid.uuid4()),
        'sender_id': sender_id,
        'format': 'html',
        'subject': tmpl['subject'],
        'message': tmpl['html'],
        'delay_id': '1_hour',
        'template_vars': tmpl['vars'],
        'metadata': {
            'template_id': tmpl['id'],
        },
    }
    if params:
        email.update(params)

    return email


def fetch_template(tmpl_id):
    '''fetch: fetch an email template
    '''
    if not tmpl_id.endswith('.html'):
        tmpl_id += '.html'

    resp = api.get('/email-templates/{}'.format(tmpl_id))
    return fmt(tmpl_id, resp)
