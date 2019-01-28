from flask import Flask
from flask import request
import socket
import os

app = Flask(__name__)
 
 
@app.route('/')
def hello_whale():
    os_env = os.environ
    r_ip = request.environ.get('HTTP_X_REAL_IP', request.remote_addr)
    req_env = request.environ
    hostname = socket.gethostname()
    return "<h2>Hello there!</h2> <br><br><b>hostname/container-id:</b> %s<br><br>OS Environment Vars:<br>%s<br><br>Request Environment Vars:<br>%s<br><br>Client IP: %s"%(hostname, os_env, req_env, r_ip)


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0')
