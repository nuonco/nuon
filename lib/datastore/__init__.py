import collections
import datetime
import json
import logging
import os
import time

from proto.datetime_helpers import DatetimeWithNanoseconds
from google.cloud import datastore as gcp_datastore
from google.oauth2 import service_account
from google.cloud import exceptions
import google.api_core.exceptions

from lib import log
from lib import config


SCOPES = (
    'https://www.googleapis.com/auth/datastore',
)


Entity = gcp_datastore.Entity


def to_dict(val):
    if isinstance(val, DatetimeWithNanoseconds):
        return datetime.datetime.fromtimestamp(int(time.mktime(val.timetuple())))

    if isinstance(val, gcp_datastore.Entity):
        val = dict(val)
        for k in val.keys():
            v = val[k]
            val[k] = to_dict(v)

        return val

    if isinstance(val, list):
        new_val = []
        for v in val:
            if isinstance(v, gcp_datastore.Entity):
                print('yes')
                v = to_dict(v)
            new_val.append(v)

        return new_val

    return val


DeadlineExceeded = google.api_core.exceptions.DeadlineExceeded
InvalidArgument = google.api_core.exceptions.InvalidArgument
TooManyRequests = google.api_core.exceptions.TooManyRequests
BadRequest = google.api_core.exceptions.BadRequest
GatewayTimeout = google.api_core.exceptions.GatewayTimeout
ServiceUnavailable = google.api_core.exceptions.ServiceUnavailable
GoogleAPICallError = google.api_core.exceptions.GoogleAPICallError
RetryError = google.api_core.exceptions.RetryError


def _kind(kind, prefix=None, version=None):
    '''kind: returns a datastore kind given the correct name
    '''
    cfg = config.load()

    prefix = prefix if prefix else cfg.datastore_prefix
    version = version if version else cfg.datastore_version

    kind = '{}__{}__{}'.format(prefix, kind.lower(), version)
    return kind


def _key(kind, key, prefix=None, version=None):
    '''kind: returns a fully resolved name with versioning for the given name
    '''
    return new_client().key(_kind(kind, prefix=prefix, version=version), key)


def _from_entity(entity):
    if not entity:
        return

    return entity


def new_client(_cache={}):
    '''new_client: returns a new datastore client
    '''
    if 'client' in _cache:
        return _cache['client']

    cfg = config.load()
    if isinstance(cfg.gcp_credentials_blob, dict):
        gcp_credentials = cfg.gcp_credentials_blob
    else:
        gcp_credentials = json.loads(cfg.gcp_credentials_blob)

    creds = service_account.Credentials.from_service_account_info(gcp_credentials)
    creds = creds.with_scopes(SCOPES)

    client = gcp_datastore.Client(credentials=creds, project=cfg.gcp_project_id)

    _cache['client'] = client
    return client


def put(kind, key, obj, exclude_from_indexes=[], prefix=None, version=None):
    full_key = _key(kind, key, prefix=prefix, version=version)
    entity = Entity(full_key, exclude_from_indexes=exclude_from_indexes)
    entity.update(obj)

    log.stderr('creating key: `{}.{}`...'.format(kind, key))
    client = new_client()
    client.put(entity)


def get_multi(kind, keys, prefix=None, version=None):
    '''get_multi: get multiple objects
    '''
    client = new_client()
    full_keys = []

    for key in keys:
        full_key = _key(kind, key, prefix=prefix, version=version)
        full_keys.append(full_key)

    results = client.get_multi(full_keys)
    return map(lambda o: _from_entity(o), list(results))


def put_multi(kind, key_objs, exclude_from_indexes=[], prefix=None, version=None):
    '''put_multi: accept a list of (key, obj) tuples and do a multi put
    '''
    client = new_client()
    entities = []
    for key, obj in key_objs:
        full_key = _key(kind, key, prefix=prefix, version=version)
        entity = Entity(full_key, exclude_from_indexes=exclude_from_indexes)
        entity.update(obj)
        entities.append(entity)

    client.put_multi(entities)


def chunked_put_multi(kind, key_objs, exclude_from_indexes=[], prefix=None, version=None, verbose=False,
        chunk_size=None):
    '''chunked_put_multi: accept a list of any size and chunk, calling put_multi for each chunk
    '''
    cfg = config.load()
    key_objs = list(key_objs)
    idx = 0

    def _(iterable):
        size = chunk_size or int(cfg.datastore_batch_size)
        for i in range(0, len(iterable), size):
            yield iterable[i:i + size]

    for chunk in _(key_objs):
        idx += len(chunk)
        while True:
            try:
                if verbose:
                    log.stderr('putting idx {}'.format(idx))
                put_multi(kind, chunk, exclude_from_indexes=exclude_from_indexes, prefix=prefix, version=version)
                break
            except (GoogleAPICallError, RetryError) as e:
                logging.exception('put_multi error', e)
                log.stderr('{} received sleeping 1s before trying again'.format(type(e)))
                time.sleep(1)


def fetch(kind, order=[], filters=[], limit=None, prefix=None, version=None, keys_only=False, offset=0):
    client = new_client()
    query = client.query(kind=_kind(kind, prefix=prefix, version=version))

    for _filter in filters:
        query.add_filter(*_filter)

    if keys_only:
        query.keys_only()

    if order:
        query.order = order

    return query.fetch(limit=limit, offset=offset)


def get(kind, key, prefix=None, version=None):
    client = new_client()

    ds_key = _key(kind, key, prefix=prefix, version=version)
    entity = client.get(ds_key)

    return _from_entity(entity)


def delete(kind, key, prefix=None, version=None):
    client = new_client()

    ds_key = _key(kind, key, prefix=prefix, version=version)
    client.delete(ds_key)
