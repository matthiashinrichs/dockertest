from flask import Flask
from flask import request
import socket
import os

app = Flask(__name__)
 
 
@app.route('/')
def hello_whale():
    os_env = os.environ
    r = request.environ.get('HTTP_X_REAL_IP', request.remote_addr)
    hostname = socket.gethostname()
    return "<h2>Hello there!</h2> <br><br>%s<br><br>%s<br><br>%s"%(hostname, os_env, r)


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0')
