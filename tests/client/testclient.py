import redis
import time
import sys
import os
from multiprocessing import Process
from multiprocessing import Queue
import functools

print = functools.partial(print, flush=True)

#40 min => 2400sec
DEADLINE = int(os.environ.get("DEADLINE",2400))

TEST_VALUE = os.environ.get("TEST_VALUE","testresult")

REDIS_HOST = os.environ.get("REDIS_HOST","redis-sentinel-proxy")

REDIS_PORT = os.environ.get("REDIS_PORT",6666)

REDIS_DB = os.environ.get("REDIS_DB",0)

PRINT_JUST_ROUND = int(os.environ.get("PRINT_JUST_ROUND",1000))

def connect():
    for i in range(3):
        try:
            redis_pool = redis.ConnectionPool(
                    host = REDIS_HOST, port = 6666, db = 0,
                    retry_on_timeout = True, decode_responses = True, socket_keepalive = True, 
                    socket_connect_timeout = 60, socket_timeout = 60 )

            r = redis.Redis(
                    connection_pool = redis_pool, single_connection_client=True,
                    retry_on_timeout=True, decode_responses=True, socket_keepalive=True,
                    socket_connect_timeout = 60, socket_timeout = 6 )

            r.set("test", TEST_VALUE)
            return r
        except Exception as e:
            print("test failed",e)
            return None
        #init retry after sleep
        time.sleep(3)

def worker(procnum, q):

    r = connect()
    if r is None:
        q.put(False)
        return
    overall_start = time.time()
    cnt = 0
    while True:
        start = time.time()
        try: 
            retval = r.get("test")
            diff = time.time()-start
            if diff > 1 or retval != TEST_VALUE :
                print("{0} took {1}s".format(retval,diff))
                if retval is None or retval != TEST_VALUE :
                    print("test failed")
                    q.put(False)
                    return
            if cnt % PRINT_JUST_ROUND == 0 :
                print("passed {0} rounds".format(cnt))
        except redis.exceptions.ConnectionError as e:
            if "Connection closed" in str(e):
                print("ConnectioError: {0} took {1}s".format(e,diff))
                continue
            print("test failed",e)
            q.put(False)
            return
        except Exception as e:
            print("test failed")
            print("Exception: {0} took {1}s".format(e,diff))
            q.put(False)
            return
        if int(time.time() - overall_start) > DEADLINE:
            print("deadline reached,worker {0}".format(procnum))
            return
        cnt += 1

if __name__ == '__main__':
    q = Queue(maxsize=0)
    wrkrs = []
    for i in range(10):
        p = Process(target=worker, args=(i,q))
        wrkrs.append(p)
        p.start()
    for wrk in wrkrs:
        wrk.join()
    
    while True:
        if q.empty():
            sys.exit(0)
        val = q.get()
        if val is False:
            sys.exit(1)
    sys.exit(0)
