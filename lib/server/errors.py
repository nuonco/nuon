from lib.errors import Error


class TokenConfigError(Error):
    status_code = 500

    def error(self):
        return 'unable to load auth tokens from `TOKENS` env var'


class TokenAuthError(Error):
    status_code = 401

    def error(self):
        return 'X-Auth-Token header value null or invalid'


class InvalidContentTypeError(Error):
    status_code = 415

    def __init__(self, content_type):
        self.content_type = content_type

    def error(self):
        return 'expecting content-type:application/json header received {}'.format(self.content_type)


class InvalidAcceptMimeTypesError(Error):
    status_code = 415

    def __init__(self, content_type):
        self.content_type = content_type

    def error(self):
        return 'expecting accept:application/json header received {}'.format(self.content_type)
