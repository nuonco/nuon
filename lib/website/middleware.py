import functools

import bs4


def prettify_html(fn):
    '''prettify: wrap the original function and return "beaitified html"
    '''
    @functools.wraps(fn)
    def _(*args, **kwargs):
        html = fn(*args, **kwargs)
        return bs4.BeautifulSoup(html, 'html.parser').prettify()

    return _
