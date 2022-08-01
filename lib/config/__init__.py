import __main__
import json
import os

from lib import log
from lib import errors
from lib import dictobject

if not os.environ.get('PT_ENV'):
    from google.cloud import kms_v1

import yaml


_OVERRIDES = {}

SERVICE_NAME = None


def set_service_name(svc_name):
    global SERVICE_NAME
    SERVICE_NAME = svc_name


class ConfigError(errors.Error):
    '''ConfigError: represents an error
    '''
    def __init__(self, key):
        self.key = key

    def error(self):
        return 'error: config: key `{}` not found in environment. expected `{}`'.format(self.key, self.key.upper())


def _load_key(orig_key):
    '''_load_key: loads a key from the environment and returns the value
    '''
    key = orig_key.upper()

    if key not in os.environ:
        raise ConfigError(orig_key)

    val = os.environ[key]

    if ',' in val:
        val = val.split(',')

    return val


def override(key, value):
    '''override: override a config setting
    '''
    _OVERRIDES[key] = value


def find_workspace_cfg(start_dir):
    '''find_workspace_cfg: find the workspace directory
    '''
    if start_dir == '/':
        log.stderr('unable to find workspace.yml')
        return

    filepath = os.path.join(start_dir, 'workspace.yml')
    if os.path.exists(filepath):
        return filepath

    return find_workspace_cfg(os.path.abspath(os.path.join(start_dir, '../')))


def load_kms_file(cfg_file):
    '''load_kms: load keys from local config.yml kms
    '''
    client = kms_v1.KeyManagementServiceClient()
    base_dir = os.path.dirname(__main__.__file__)

    workspace_cfg_path = find_workspace_cfg(base_dir)
    if not workspace_cfg_path:
        return

    if not os.path.exists(workspace_cfg_path):
        log.stderr('no workspace.yml file found at {}'.format(workspace_cfg_path))
        return

    with open(workspace_cfg_path, 'r') as fh:
        cfg = yaml.safe_load(fh)

    svc_name = SERVICE_NAME if SERVICE_NAME else os.environ['SERVICE']
    if not svc_name:
        log.stderr('service name not set')
        return


    key_name = client.crypto_key_path_path(cfg['kms_gcp_project_id'],
                                           cfg['kms_gcp_location_id'],
                                           cfg['kms_key_ring'],
                                           cfg['kms_key_prefix'] + '-' + svc_name)

    with open(cfg_file, 'rb') as fh:
        ciphertext = fh.read()

    response = client.decrypt(key_name, ciphertext)
    contents = response.plaintext
    return yaml.safe_load(contents)


def load_kms():
    '''load_kms: load kms secrets by resolving a shared.yml and config.yml filepath first
    '''
    resolved_cfg = {}
    base_dir = os.path.dirname(__main__.__file__)

    for filename in ('shared.yml.kms.enc', 'config.yml.kms.enc'):
        cfg_file = os.path.join(base_dir, filename)
        if not os.path.exists(cfg_file):
            continue

        cfg = load_kms_file(cfg_file)
        resolved_cfg.update(cfg)

    return resolved_cfg


def load(keys=[], _cache={}):
    '''load: loads configuration from the environment and returns a dict object
    '''
    if 'cfg' in _cache:
        return _cache['cfg']

    if os.environ.get('PT_ENV', ''):
        with open('/settings/settings.json', 'r') as fh:
            content = fh.read()

        cfg = json.loads(content)
        cfg['environment'] = os.environ.get('PT_ENV')
        cfg['git_short_ref'] = os.environ.get('PT_VERSION')
        return dictobject.AnyObject(cfg)

    cfg = load_kms()
    if not cfg:
        return

    env_keys = ('ENVIRONMENT', 'GIT_SHORT_REF')
    for key in env_keys:
        cfg[key.lower()] = os.environ.get(key, '')

    cfg.update(_OVERRIDES)
    cfg = dictobject.AnyObject(cfg)

    _cache['cfg'] = cfg
    return cfg
