from flask import Flask
from flask import request
import socket
import os

app = Flask(__name__)
 
 
@app.route('/')
def hello_whale():
    os_env = os.environ
    r = request.remote_addr
    hostname = socket.gethostname()
    return "Whale, Hello there! <br>%s<br><br>%s<br><br>%s"%(hostname, os_env, r)


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0')
