// import { error } from '@sveltejs/kit';

const base = 'http://127.0.0.1:8090/api/v1';

async function send({ method, path, data, token }) {
	const opts = { method, headers: {} };

	if (data) {
		opts.headers['Content-Type'] = 'application/json';
		opts.body = JSON.stringify(data);
	}

	if (token) {
		opts.headers['X-Auth-Token'] = token;
	}

    const res = fetch(`${base}/${path}`, opts)
		.then((response) => response.json())
		.then((response) => {
			return response;
		})
		.catch((err) => {
            return err;
		});

	return res;
}

export function get(path, token) {
	return send({ method: 'GET', path, token });
}

export function del(path, token) {
	return send({ method: 'DELETE', path, token });
}

export function post(path, data, token) {
	return send({ method: 'POST', path, data, token });
}

export function put(path, data, token) {
	return send({ method: 'PUT', path, data, token });
}
