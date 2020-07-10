''' test an mtconnect adapter by reading from/ writing to tcp socket
'''
import argparse
import asyncio
import logging


async def main(host, port):
    ''' connect to adapter and maintain connection until interrupted
    '''
    logging.debug('Connecting...')
    reader, writer = await asyncio.open_connection(host, port)

    try:
        await asyncio.gather(read_from(reader), ping_to(writer))
    finally:
        writer.close()
        await writer.wait_closed()


async def ping_to(writer, interval=10):
    ''' write mandatory ping message at interval to adapter
    '''
    message = '* PING'
    while True:
        logging.debug('Sending ping.')
        writer.write(message.encode())
        await writer.drain()
        await asyncio.sleep(interval)


async def read_from(reader):
    ''' read newline delimited bytes from adapter
    '''
    logging.debug('Reading...')
    async for data in reader:
        line = data.decode().strip()
        if line.startswith('* PONG'):
            logging.debug('Received pong.')
            continue
        logging.info('Received: %s', line)


if __name__ == '__main__':
    logging.basicConfig(format='%(asctime)s %(levelname)s %(message)s', level=logging.DEBUG)
    parser = argparse.ArgumentParser(description='Test tcp connection to MTConnect adapter.')
    parser.add_argument('--host', metavar='host', type=str, nargs='?',
                        help='target host', required=True)
    parser.add_argument('--port', metavar='123', type=int, help='target port',
                        default=7878)

    args = parser.parse_args()
    logging.debug('Using host=%s and port=%d', args.host, args.port)

    try:
        loop = asyncio.get_event_loop()
        loop.run_until_complete(main(args.host, args.port))
    except KeyboardInterrupt:
        pass
