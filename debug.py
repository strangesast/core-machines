''' test an mtconnect adapter by reading from/ writing to tcp socket
'''
import argparse
import asyncio
import logging
from datetime import datetime
from itertools import zip_longest
from collections import defaultdict
from pprint import pprint


async def main(host, port, raw = False):
    ''' connect to adapter and maintain connection until interrupted
    '''
    logging.debug('Connecting...')
    reader, writer = await asyncio.open_connection(host, port)

    task, pings, lines = splitter(reader)
    try:
        tasks = [task, read_from(lines, raw), ping_to(pings, writer)]
        await asyncio.gather(*tasks)
    finally:
        writer.close()
        await writer.wait_closed()


def splitter(inp):
    q1, q2 = asyncio.Queue(), asyncio.Queue()
    prefix = '* PONG'

    async def fun(source):
        try:
            async for data in source:
                line = data.decode().strip()
                if line.startswith(prefix):
                    await q1.put(line[len(prefix):])
                else:
                    await q2.put(line)
        except asyncio.CancelledError:
            pass

    return fun(inp), q1, q2


async def ping_to(source, sink, default_interval=10):
    ''' write mandatory ping message at interval to adapter
        client is disconnected if ping not sent 2x interval
    '''
    ping = '* PING'.encode()
    try:
        while True:
            logging.debug('Sending ping.')
            sink.write(ping)
            await sink.drain()
            pong = await source.get()
            try:
                interval = int(pong.strip()) / 1000
                assert interval > 0, 'invalid interval'
            except:
                interval = default_interval

            logging.debug('Got pong! (Interval {})'.format(interval))
            await asyncio.sleep(interval)
            source.task_done()
    except asyncio.CancelledError:
        pass


NUMERIC = [
        'part_count',
        'SspeedOvr',
        'Fovr',
        'tool_id',
        'line',
        'path_feedrate',
        'f_command',
        'Aact',
        'Xact',
        'Yact',
        'Zact',
        'Cact',
        'Aload',
        'Xload',
        'Yload',
        'Zload',
        'Cload',
        'S1speed',
        'S1load',
        'S2speed',
        'S2load']


def parse(k, s):
    if k in NUMERIC:
        return float(s)
    if k == 'path_position':
        return [float(s) for s in s.split(' ')]
    return s


async def read_from(source, raw):
    ''' read newline delimited bytes from adapter
    '''
    logging.debug('Reading...')
    state = {}
    try:
        while True:
            line = await source.get()
            if raw:
                logging.info('Received: %s', line)
            else:
                # TODO: handle escaped pipes
                parts = [s.strip() for s in line.split('|')]
                ts, parts = parts[0], parts[1:]
                ts = datetime.strptime(ts, '%Y-%m-%dT%H:%M:%S.%fZ')

                parts = zip_longest(parts[0::2], parts[1::2]) if parts[0] != 'message' else [(parts[0], parts[2])]

                obj = [(k[0:-1] if k.endswith('2') else k, parse(k, v)) for k, v in parts if k and v not in ['UNAVAILABLE', '']]

                if len(obj):
                    state.update(dict(obj))

                pprint(state)

            source.task_done()
    except asyncio.CancelledError:
        pass


if __name__ == '__main__':
    logging.basicConfig(format='%(asctime)s %(levelname)s %(message)s', level=logging.INFO)
    parser = argparse.ArgumentParser(description='Test tcp connection to MTConnect adapter.')
    parser.add_argument('--host', metavar='host', type=str, nargs='?',
                        help='target host', required=True)
    parser.add_argument('--port', metavar='123', type=int, help='target port',
                        default=7878)
    parser.add_argument('--raw', help='leave raw output', dest='raw', action='store_true')

    args = parser.parse_args()
    logging.debug('Using host=%s and port=%d', args.host, args.port)

    try:
        loop = asyncio.get_event_loop()
        loop.run_until_complete(main(args.host, args.port, args.raw))
    except KeyboardInterrupt:
        pass
