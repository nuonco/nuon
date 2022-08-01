import datetime
import time

from lib import config
from lib import obj
from lib import log
from shared import errs
from shared import params

import email_validator
import yaml
import sendgrid
from sendgrid.helpers.mail import *


PREFIX = 'growth-tool'


FIELDS = [
    'subject',
    'message',
    'format',
    'recipient_email',
]

OPTIONAL_FIELDS = [
    'metadata',
    'template',
    'recipient_name',
]

obj.EXCLUDE_FROM_INDEXES.append('message')


def get_templates():
    '''get_templates: return a list of all templates
    '''
    return []


def get_delays():
    '''get_delays: return a list of delays
    '''
    delays = [
        ('1_hour', datetime.timedelta(hours=1)),
        ('no_delay', None),
        ('5_min', datetime.timedelta(minutes=5)),
        ('2_hour', datetime.timedelta(hours=2)),
        ('12_hour', datetime.timedelta(hours=12)),
        ('1_day', datetime.timedelta(days=1)),
        ('2_day', datetime.timedelta(days=7)),
    ]

    return [{'id': d[0], 'delta': d[1]} for d in delays]


def get_senders():
    '''get_senders: get senders
    '''
    senders = [
        ('jon_powertoolsdev', 'jon@powertools.dev', 'Jon Morehouse'),
        ('zain_powertoolsdev', 'zain@powertools.dev', 'Zain Ahmed'),

        ('jon_powertoolsdevinfo', 'jon@powertoolsdev.info', 'Jon Morehouse'),
        ('zain_powertoolsdevinfo', 'zain@powertoolsdev.info', 'Zain Ahmed'),

        ('team_powertoolsdev', 'team@powertools.dev', 'Team PowerTools'),
        ('accounts_powertoolsdev', 'accounts@powertools.dev', 'PowerTools Accounts'),
    ]
    objs = []
    for sender in senders:
        objs.append({
            'id': sender[0],
            'email': sender[1],
            'name': sender[2],
        })

    return objs


def get_formats():
    '''get_formats: return the formats
    '''
    return [
        'github_markdown',
        'markdown',
        'plaintext',
        'html',
        'amp',
    ]


def format_email(email):
    '''format_email: format the provided email
    '''
    try:
        e = email_validator.validate_email(email['recipient_email'])
        email['recipient_address'] = e.email
        email['recipient_email_status'] = 'ok'
    except Exception as e:
        email['recipient_email_status'] = 'invalid'
        email['recipient_email_error'] = str(e)

    if email['format'] == 'github_markdown':
        email['message'] = {
            'raw': email['message'],
            'rendered': obj.render_markdown(email['message'], api=True)['html'],
        }
    elif email['format'] == 'markdown':
        email['message'] = {
            'raw': email['message'],
            'rendered': obj.render_markdown(email['message'], api=False)['html'],
        }
    else:
        email['message'] = {
            'raw': email['message'],
            'rendered': email['message'],
        }

    email['sender'] = [s for s in get_senders() if s['id'] == email['sender_id']][0]

    if email['delay_id']:
        email['delay'] = [d for d in get_delays() if d['id'] == email['delay_id']][0]

    return email


def fetch(email_id):
    '''fetch: fetch an email
    '''
    email = obj.fetch('email', email_id, prefix=PREFIX)
    if not email:
        return

    return format_email(email)


def parse(data):
    '''parse_data: parse the provided data into the correct object
    '''
    params.ensure(data, FIELDS)

    if isinstance(data.get('metadata'), str):
        data['metadata'] = yaml.safe_load(data['metadata'])

    return data


def save(email_id, data):
    '''save: save an email
    '''
    data = parse(data)

    data['recipient_email'] = email_validator.validate_email(data['recipient_email']).email

    if not fetch(email_id):
        obj.create('email', email_id, data, prefix=PREFIX, exclude_from_indexes=('message',))
        return

    obj.update('email', email_id, data, prefix=PREFIX, exclude_from_indexes=('message',))


def cancel(email_id):
    '''cancel: cancel an email through the sendgrid api
    '''
    cfg = config.load()
    email = fetch(email_id)

    # TODO: update sendgrid
    sg = sendgrid.SendGridAPIClient(cfg.sendgrid_api_key)
    response = sg.client.user.credits.get()
    print(response)


def schedule(email_id):
    '''schedule: schedule the email to be sent
    '''
    cfg = config.load()

    email = fetch(email_id)

    sender = Email(email['sender']['email'], email['sender']['name'])
    recipient = To(email['recipient_email'], email.get('recipient_name', ''))

    content_type = 'text/plain' if email['format'] == 'plaintext' else 'text/html'

    mail = Mail(from_email=sender,
                to_emails=[recipient],
                reply_to=sender,
                subject=email['subject'])
    mail.content = Content(content_type, email['message']['rendered'])
    mail.custom_arg = CustomArg('pt_email_id', email_id)

    ts = int(time.time())
    if email.get('delay') and email['delay'].get('delta'):
        send_at_ts = ts + int(email['delay']['delta'].total_seconds())
        mail.send_at = SendAt(send_at_ts)
        email.pop('delay')

    body = mail.get()

    sendgrid_api = sendgrid.SendGridAPIClient(cfg.sendgrid_api_key)
    resp = sendgrid_api.client.mail.send.post(request_body=body)

    if resp.status_code != 202:
        raise Exception('unable to send message {} {}'.format(resp.status_code, resp.body))

    message_id = resp.headers['X-Message-Id']
    email['message'] = email['message']['raw']
    email['sendgrid_message_id'] = message_id

    obj.update('email', email_id, email, prefix=PREFIX)


def fetch_all(filters, limit=None, ignore=True):
    '''fetch_all: fetch all emails
    '''
    # NOTE: in order to avoid needing to build a bunch of indexes, we apply most filters locally
    qfilters = []
    if 'start_ts' in filters:
        qfilters.append(['created_at', '>', filters.pop('start_ts')])

    emails = obj.fetch_all('email', order='-created_at', limit=limit, filters=qfilters, prefix=PREFIX)

    # manually filter emails locally
    filtered = []
    for email in emails:
        for k, v in filters.items():
            if not v:
                continue

            # parse recipient
            if ignore and email['recipient_email'].endswith('@powertools.dev'):
                continue

            if email.get(k) != v:
                continue
            filtered.append(email)
            break

    return filtered
