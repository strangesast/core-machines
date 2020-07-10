import asyncio
import motor.motor_asyncio
from pprint import pprint


base_pipeline = [
  {'$addFields': {'timestamp': {'$dateFromString': {'dateString': '$timestamp'}}}},
  {'$addFields': {'values': {'$cond': {
      'if': {'$eq': [{'$arrayElemAt': ['$values', 0]}, 'message']},
      'then': ['message', {'$arrayElemAt': ['$values', 2]}],
      'else': '$values'
  }}}},
  {'$addFields': {'values': {'$map': {
      'input': {'$range': [0, {'$size': '$values'}, 2]},
      'in': {'$slice': ['$values', '$$this', 2]},
  }}}},
  {'$addFields': {'values': {'$filter': {
      'input': '$values',
      'cond': {'$ne': [{'$arrayElemAt': ['$$this', 0]}, ""]},
  }}}},
  {'$unwind': '$values'},
  {'$project': {
      '_id': 1,
      'timestamp': 1,
      'key': {'$arrayElemAt': ['$values', 0]},
      'value': {'$arrayElemAt': ['$values', 1]},
  }},
]

async def main():
    uri = 'mongodb://localhost:27017'
    client = motor.motor_asyncio.AsyncIOMotorClient(uri)

    db = client.monitoring

    ignore = ['Ztravel', 'Zact', 'Xoverheat', 'S1load', 'S1servo', 'Zservo', 'Cload', 'S2load', 'block', 'S1speed', 'Xservo', 'Zload', 'S2speed', 'Fovr', 'Xtravel', 'servo', 'S2servo', 'path_position', 'SspeedOvr', 'Cservo', 'active_axes', 'Coverheat', 'Xact', 'path_feedrate', 'Ctravel', 'Xload', 'Zoverheat', 'Cact']

    #'mode', '197', 'execution', 'line', 'message', 'program_comment', 'f_command', 'part_count', 'estop', '3055', 'avail', 'comms', 'program',
    interesting = ['tool_id', 'system']

    unique = set()
    l = []
    with open('output.txt', 'w') as output:
        async for doc in db['input'].aggregate([*base_pipeline, {'$match': {'$expr': {'$not': {'$in': ['$key', ignore]}}}}]):
            key = doc['key']
            unique.add(key)
            #if key not in ignore:
            output.write(f'{key.ljust(16)} {doc["timestamp"]} {doc["value"]}\n')
    print(unique)

if __name__ == '__main__':
    asyncio.run(main())
