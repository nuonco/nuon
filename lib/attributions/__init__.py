import base64
import urllib.parse
import uuid

from lib.obj import params
from lib import obj


DATASTORE_PREFIX = 'attributions'
DATASTORE_VERSION = 'v1'


def tag_url(url, tag):
    '''tag_url: add a `ptr=tag` query parameter to the url and return it
    '''
    parts = list(urllib.parse.urlparse(url))
    query = dict(urllib.parse.parse_qs(parts[4], keep_blank_values=True))
    query['ptr'] = tag
    parts[4] = urllib.parse.urlencode(dict(query), doseq=True)
    new_url = urllib.parse.urlunparse(parts)

    return new_url


def create_website_url(data, write_later=False):
    '''create_website_url: create a website url
    '''
    params.ensure(data, ['url', 'email'])

    url = {
        'url': data['url'],
        'created_by': data['email'],
        'updated_by': data['email'],
    }
    if 'metadata' in data:
        url.update(data['metadata'])

    url.update(data['metadata'])

    url['id'] = data.get('id') or obj.create_id('w')

    url['tag'] = url['id']
    url['attributed_url'] = tag_url(url['url'], url['tag'])

    if write_later:
        return url

    obj.create('website-url', url['id'], url, prefix=DATASTORE_PREFIX, version=DATASTORE_VERSION)
    return url


def create_urls(urls):
    '''create_urls: persist many urls that were created using the `write_later` flag on the creator
    '''
    key_objs = []
    for url in urls:
        key_objs.append((url['id'], url))

    obj.create_all('website-url', key_objs, prefix=DATASTORE_PREFIX, version=DATASTORE_VERSION)


def create_redirect_url():
    '''create_redirect_url: create a url that can be redirected through our redirect service
    '''


def new_website_session(params):
    '''new_website_session: create a new website session, given the parameters. If no attribution tag exists, make an
    organic session.

    These sessions are not strongly connected, notably they don't do two things

    * track expirations
    * check if the ptr key is actually attributed

    The reason for this is to simplify the implementation and enable us to dynamically create attribution urls in the
    future without storing the extra data if we don't want.
    '''
    session = {
        'id': str(uuid.uuid4()),
        'organic': 'ptr' in params,
        'tag': params['ptr'] if 'ptr' in params else 'o',
    }
    session.update(params)

    raw_key = '{}|{}'.format(session['id'], session['tag'])
    raw_key = raw_key.encode('ascii')
    key = base64.b64encode(raw_key).decode('ascii')
    key = key.replace('=', '', -1)
    session['key'] = key

    obj.create('website-session', key, session, prefix=DATASTORE_PREFIX, version=DATASTORE_VERSION)
    return session


def create_website_event(key, params):
    '''create_website_event: create an event
    '''
    if not key.endswith('=='):
        key = '{}=='.format(key)

    key = base64.b64decode(key)
    key = key.decode('ascii')
    session_id, tag = key.split('|')

    event = {}
    event.update(params)
    event['session_id'] = session_id
    event['tag'] = tag
    event['id'] = str(uuid.uuid4())

    index_keys = ['tag', 'session_id', 'created_at']
    excluded = []
    for key in event.keys():
        if key not in index_keys:
            excluded.append(key)

    obj.create('website-event', event['id'], event,
               exclude_from_indexes=excluded,
               prefix=DATASTORE_PREFIX,
               version=DATASTORE_VERSION)


def fetch_website_url(tag):
    return obj.fetch_one('website-url', filters=[['tag', '=', tag]],
                         prefix=DATASTORE_PREFIX,
                         version=DATASTORE_VERSION)


def fetch_website_activity(tag):
    '''fetch_website_activity: fetch website activity
    '''
    activity = {}

    events = obj.fetch_all('website-event', filters=[['tag', '=', tag]],
                           prefix=DATASTORE_PREFIX,
                           version=DATASTORE_VERSION)
    activity['events'] = events

    sessions = obj.fetch_all('website-session', filters=[['tag', '=', tag]],
                             prefix=DATASTORE_PREFIX,
                             version=DATASTORE_VERSION)
    activity['sessions'] = sessions

    return activity
