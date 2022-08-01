import datetime
import hashlib
import json
import urllib

import markdown2
import pytz
import requests

from lib import config
from lib import log
from lib import datastore


EXCLUDE_FROM_INDEXES = []


def format_id(val):
    '''format_id: formats the value into an id
    '''
    val = val.replace(' ', '_')
    val = val.lower()

    return val


def create_short_code(offset):
    '''create_short_code: create a friendly short code
    '''
    alphabet = '23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ'
    if offset == 0:
        return alphabet[0]

    arr = []
    base = len(alphabet)
    while offset:
        offset, rem = divmod(offset, base)
        arr.append(alphabet[rem])
    arr.reverse()
    return ''.join(arr)


def create_id(ns='t', prefix=None, version=None, default_offset=100000):
    '''create_id: create an id by using an auto incrementing integer to create a friendly id
    '''
    ids = create_ids(cnt=1, ns=ns, prefix=prefix, version=version, default_offset=default_offset)
    return ids[0]


def create_ids(cnt, ns='t', prefix=None, version=None, default_offset=100000):
    '''create_ids: create multiple ids
    '''
    client = datastore.new_client()

    with client.transaction():
        key = datastore._key('id-offset', ns, prefix=prefix, version=version)
        remote_offset = client.get(key)
        offset = remote_offset['offset'] if remote_offset else default_offset

        offset += cnt
        entity = datastore.Entity(key)
        entity.update({
            'created_at': datetime.datetime.now(pytz.UTC),
            'offset': offset,
        })

        client.put(entity)

    ids = []

    for i in range(cnt):
        id_offset = offset + i

        # create a friendly hash of this value
        short_code = create_short_code(id_offset)
        new_id = '{}{}'.format(ns, short_code)
        ids.append(new_id)

    return ids


def create(entity, obj_id, obj, exclude_from_indexes=None, prefix=None, version=None):
    '''create: create the given object
    '''
    obj['id'] = obj_id
    obj['created_at'] = datetime.datetime.now(pytz.UTC)
    obj['updated_at'] = datetime.datetime.now(pytz.UTC)

    exclude = exclude_from_indexes or EXCLUDE_FROM_INDEXES

    datastore.put(entity, obj_id, obj, exclude_from_indexes=exclude, prefix=prefix, version=version)
    return obj


def create_all(entity, key_objs, exclude_from_indexes=None, prefix=None, version=None):
    '''create_all: create all objects in the list of key objs
    '''
    for key, obj in key_objs:
        obj['created_at'] = datetime.datetime.now(pytz.UTC)
        obj['updated_at'] = datetime.datetime.now(pytz.UTC)

    exclude = exclude_from_indexes or EXCLUDE_FROM_INDEXES
    datastore.chunked_put_multi(entity, key_objs, exclude_from_indexes=exclude, prefix=prefix, version=version)


def update(entity, obj_id, obj, exclude_from_indexes=None, prefix=None, version=None):
    '''update: update the given object
    '''
    obj['updated_at'] = datetime.datetime.now(pytz.UTC)

    exclude = exclude_from_indexes or EXCLUDE_FROM_INDEXES
    datastore.put(entity, obj_id, obj, exclude_from_indexes=exclude, prefix=prefix, version=version)


def upsert(entity, obj_id, obj, **kwargs):
    '''upsert: upsert the given object by creating a new one if it does not exist, or patching an existing one
    '''
    old_obj = fetch(entity, obj_id, **kwargs)
    if not old_obj:
        create(entity, obj_id, obj, **kwargs)
        return

    old_obj.update(obj)
    update(entity, obj_id, old_obj, **kwargs)


def fetch_all(entity, order=None, filters=[], limit=None, keys_only=False, prefix=None, version=None):
    '''fetch_all: fetch all objects of the entity
    '''
    objs = datastore.fetch(entity,
                           limit=limit,
                           order=order,
                           filters=filters,
                           keys_only=keys_only,
                           prefix=prefix,
                           version=version)
    return list(objs)


def fetch_latest(entity, filters=None, order=None, limit=None, prefix=None, version=None):
    '''fetch_latest: return the latest item to match the filters, by loading all items and sorting locally
    '''
    objs = fetch_all(entity, filters=filters, prefix=prefix, version=version)
    objs = list(objs)
    if not objs:
        return

    if len(objs) == 1:
        return objs[0]

    objs = sorted(objs, key=lambda o: o['created_at'])
    objs = list(reversed(objs))
    return objs[0]


def to_batch(items, limit):
    '''to_batch: yield a list of batches
    '''
    batch = []

    for item in items:
        batch.append(item)

        if len(batch) < limit:
            continue

        yield batch
        batch = []

    if not batch:
        return

    yield batch


def fetch_multi(entity, keys, prefix=None, version=None, to_dict=False):
    '''fetch_multi: fetch multiple objects in the namespace
    '''
    all_objs = []

    for batch in to_batch(keys, 1000):
        objs = datastore.get_multi(entity, batch, prefix=prefix, version=version)
        objs = list(objs)

        all_objs.extend(objs)

    if not to_dict:
        return all_objs

    return dict([[obj.key.name, obj] for obj in all_objs])


def fetch_all_sorted(entity, filters=None, order=None, limit=None, prefix=None, version=None, sort_key='created_at'):
    '''fetch_all_sorted: fetch all for a filter and sort locally
    '''
    objs = fetch_all(entity, filters=filters or [], prefix=prefix, version=version)
    objs = list(objs)
    if not objs:
        return []

    if len(objs) == 1:
        return [objs[0]]

    objs = sorted(objs, key=lambda o: o['created_at'])
    objs = list(reversed(objs))
    if limit:
        objs = objs[:limit]

    return objs


def fetch_one(entity, filters=None, order=None, limit=None, prefix=None, version=None):
    '''fetch: return the given object of the given entity
    '''
    objs = datastore.fetch(entity, filters=filters, order=order, limit=limit, prefix=prefix, version=version)
    objs = list(objs)
    if not objs:
        return

    return objs[0]


def fetch(entity, id, prefix=None, version=None):
    return datastore.get(entity, id, prefix=prefix, version=version)


def delete(entity, obj_id, prefix=None, version=None):
    '''delete: delete the object
    '''
    datastore.delete(entity, obj_id, prefix=prefix, version=version)


def delete_filter(entity, filters=[], prefix=None, version=None):
    keys = fetch_all(entity, filters=filters, keys_only=True, prefix=prefix, version=version)
    for key in keys:
        delete(entity, key.key.name, prefix=prefix, version=version)


def delete_all(entity, prefix=None, version=None, verbose=True):
    while True:
        keys = fetch_all(entity, prefix=prefix, version=version, keys_only=True)
        if not keys:
            break

        if verbose:
            log.stderr('deleting {} keys'.format(len(keys)))

        for key in keys:
            delete(entity, key.key.name, prefix=prefix, version=version)


def _github_request(uri, method='GET', body=None, expect_json=True, extra_headers={}):
    '''_github_fetch: make a request to the github api, returning the results
    or printing the error and exiting
    '''
    cfg = config.load()

    url = urllib.parse.urljoin('https://api.github.com/', uri)
    timeout = 2.50
    headers = {
        'Authorization': 'token {}'.format(cfg.github_api_token),
    }
    headers.update(extra_headers)

    req_fn = {
        'GET': lambda: requests.get(url, headers=headers, timeout=timeout),
        'POST': lambda: requests.post(url, headers=headers, data=json.dumps(body), timeout=timeout),
        'PATCH': lambda: requests.patch(url, headers=headers, data=json.dumps(body), timeout=timeout),
        'PUT': lambda: requests.put(url, headers=headers, timeout=timeout),
        'DELETE': lambda: requests.delete(url, headers=headers, timeout=timeout),
    }[method]

    try:
        resp = req_fn()
    except requests.exceptions.Timeout:
        raise GithubAPIError(uri, 500, 'timeout')

    if resp.status_code not in (200, 201, 204):
        log.stderr('invalid github status code {} response {}...'.format(resp.status_code, resp.text))

    if not expect_json:
        return resp.text

    return resp.json()


def create_sha256(val):
    hash_obj = hashlib.sha256(val.encode('utf-8'))
    digest = hash_obj.hexdigest()
    key = str(digest)

    return key


def render_markdown(markdown, api=True):
    '''render_markdown: render markdown into gfm html
    '''
    key = create_sha256(markdown)
    if not api:
        return {
            'markdown': markdown,
            'id': key,
            'html': markdown2.markdown(markdown),
        }

    obj = fetch('markdown', key)
    if obj:
        return obj

    body = {
        'text': markdown,
        'mode': 'gfm',
        'context': 'powertoolsdev/workspace',
    }
    resp = _github_request('/markdown', method='POST', body=body, expect_json=False)
    obj = {
        'markdown': markdown,
        'id': key,
        'html': resp,
    }
    create('markdown', key, obj, exclude_from_indexes=['html'])
    return obj


def render_markdown_many(objs, key):
    objs_by_key = {}
    for obj in objs:
        sha = create_sha256(obj[key])
        objs_by_key[sha] = obj

    cached = datastore.get_multi('markdown', objs_by_key.keys())
    cached = list(cached)
    for val in cached:
        obj = objs_by_key.pop(val.key.name)
        obj[key] = val

    for obj in objs_by_key.values():
        obj[key] = render_markdown(obj[key])

    return objs
