# Data Structures & Types (taken from MTConnect docs)

Samples:
A point-in-time measurement of a data item that is continuously changing.

Events:
Discrete changes in state that can have no intermediate value. They indicate the  
state of a specific attribute of a component. (See Events in Part 3)

Condition:
A piece of information the device provides as an indicator of its health and  
ability to function. A condition can be one of Normal, Warning, Fault, or  
Unavailable. A single condition type can have multiple Faults or Warnings at  
any given time. This behavior is different from Events and Samples where a data  
item MUST only have a single value at a given time. (See Condition in Part 3).

# Processing
1. Get `/current`.  Save events, samples, condition of each component.  
  Retrieve lastSequence
2. Get `/sample` where `from=lastSequence`
3. Recurse `/sample` with lastSequence OR use `interval=0`
