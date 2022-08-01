import contextlib
import logging
import os
import time

from lib import config
from lib import log


ENABLED = False


def serialize_tags(tags):
    '''serialize_tags: serialize a dict of tags into a list of colon separated strings.
    '''
    return ['{}:{}'.format(k, v) for k, v in tags.items() if v is not None]


def _send(_method, _key, _value=None, sample_rate=1.0, **kwargs):
    '''_send: send metrics to the underlying statsd connection or stdout
    '''
    if not ENABLED:
        return

    cfg = config.load()

    resolved_tags = {}

    resolved_tags.update(kwargs.pop('tags', {}))
    resolved_tags.update(kwargs)

    tags = serialize_tags(resolved_tags)
    if not cfg.get('write_metrics'):
        if isinstance(_value, float):
            _value = format(_value, '.2f')

        joined_tags = ' '.join('tag=%s' % tag for tag in tags)
        logging.info('metric:{}:{}={} {}'.format(_method, _key, _value, joined_tags))
        return

    method = getattr(dogstatsd, _method)
    method(_key, _value, sample_rate=sample_rate, tags=tags)


def histogram(_key, _count, **kwargs):
    '''histogram: write a histogram metric to statsd; appending any key=value tags given
    '''
    _send('histogram', _key, _count, **kwargs)


def incr(_key, _count=1, **kwargs):
    '''incr: write an incr metric to statsd; appending any key=value tags given
    '''
    _send('increment', _key, _count, **kwargs)


def gauge(_key, _count, **kwargs):
    '''gauge: write a gauge metric to statsd; appending any key=value tags given
    '''
    _send('gauge', _key, _count, **kwargs)


def timing(_key, _timing, **kwargs):
    '''timing: write a timing metric to statsd; appending any key=value tags given
    '''
    _send('timing', _key, _timing, **kwargs)


@contextlib.contextmanager
def timer(_key, **kwargs):
    '''timer: write a timing metric to statsd, based upon the context manager
    derived execution time; appending any key=value tags given.
    '''
    start = time.time()
    yield
    value = (time.time() - start) * 1000
    timing(_key, value, **kwargs)
