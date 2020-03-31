from flask import Flask
from flask import render_template
from flask import request
import socket
import sys
import os

app = Flask(__name__)
 
 
@app.route('/')
def hello_index():
    os_env = os.environ
    build_number = os.environ.get('BUILD_NUMBER', '0')
    python_version = sys.version
    r_ip = request.environ.get('HTTP_X_REAL_IP', request.remote_addr)
    x_forwarded_ip = request.environ.get('HTTP_X_ORIGINAL_FORWARDED_FOR', '0')
    req_env = request.environ
    hostname = socket.gethostname()
    css_version = os.path.getmtime("static/css/style.css")
    data = {
        'python_version': python_version,
        'os_env': os_env,
        'client_ip': r_ip,
        'request_env': req_env,
        'x_forwarded_ip': x_forwarded_ip,
        'hostname': hostname,
        'build_number': build_number,
        'css_version': css_version
    }
    return render_template('index.html', title='Home', data=data)


if __name__ == '__main__':
    app.run(debug=False, host='0.0.0.0')
