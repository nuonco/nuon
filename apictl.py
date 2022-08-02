#!/usr/bin/env python3

from lib import log

import click
import api


@click.group()
def cli():
    pass


@cli.command(name='query')
@click.argument('query', type=str)
@click.option('--env', type=str, help='use environment', default=api.DEFAULT_ENV, envvar='APICTL_ENV')
def query(query, **kwargs):
    '''query: executes a query
    '''
    if not query:
        log.stder('please pass in the name of a query')
        return
    if query not in api.all_queries_by_name():
        log.stderr('query {} not found'.format(query))
        return

    api.exec_request('query', query, kwargs['env'])


@cli.command(name='mutation')
@click.argument('mutation', type=str)
@click.option('--env', type=str, help='use environment', default=api.DEFAULT_ENV, envvar='APICTL_ENV')
def mutation(mutation, **kwargs):
    '''mutation: executes a mutation
    '''
    if not mutation:
        log.stderr('please pass in the name of a mutation')
        return
    if mutation not in api.all_mutations_by_name():
        log.stderr('mutation {} not found'.format(mutation))
        return

    api.exec_request('mutation', mutation, kwargs['env'])

@cli.command(name='list')
def list(**kwargs):
    '''list: list all queries and mutations
    '''
    for query in api.all_queries_by_name():
        log.stderr('- {} (query)'.format(query))

    for mutation in api.all_mutations_by_name():
        log.stderr('- {} (mutation)'.format(mutation))


if __name__ == '__main__':
    log.set_option('prefix', False)
    cli()
