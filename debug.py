import argparse
import asyncio
import logging


async def main(host: str, port: int):
    logging.debug('Connecting...')
    reader, writer = await asyncio.open_connection(host, port)

    try:
        await asyncio.gather(read_from(reader), ping_to(writer))
    finally:
        writer.close()
        await writer.wait_closed()


async def ping_to(writer, interval=10):
    message = '* PING'
    while True:
        logging.debug('Sending ping.')
        writer.write(message.encode())
        await writer.drain()
        await asyncio.sleep(interval)


async def read_from(reader):
    logging.debug('Reading...')
    async for data in reader:
        s = data.decode().strip()
        if s.startswith('* PONG'):
            logging.debug('Received pong.')
            continue
        logging.info(f'Received: {s}')


if __name__  == '__main__':
    logging.basicConfig(format='%(asctime)s %(levelname)s %(message)s', level=logging.DEBUG)
    parser = argparse.ArgumentParser(description='Test tcp connection to MTConnect adapter.')
    parser.add_argument('--host', metavar='host', type=str, nargs='?',
            help='target host', required=True)
    parser.add_argument('--port', metavar='123', type=int, help='target port',
            default=7878)
    
    args = parser.parse_args()
    host, port = args.host, args.port
    logging.debug(f'Using {host=} & {port=}')
    
    try:
        asyncio.run(main(host, port))
    except KeyboardInterrupt:
        pass
