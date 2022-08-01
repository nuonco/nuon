class Error(Exception):
    '''Error is a class that is used as the core of all errors, it includes a
    status code so that it's easily serializable in api responses
    '''
    status_code = 500

    def __str__(self):
        return '{}'.format(self.error())

    def __repr__(self):
        return '{}'.format(self.error())

    def error(self):
        assert False, 'not implemented'
