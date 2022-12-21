import { fail, redirect } from '@sveltejs/kit';
import * as api from '$lib/api.js';

/** @type {import('./$types').PageServerLoad} */
export async function load({ parent }) {
	const { user } = await parent();
	if (user) throw redirect(307, '/');
}

/** @type {import('./$types').Actions} */
export const actions = {
	default: async ({ request }) => {
		const data = await request.formData();

		const response = await api.post('users', {
			first_name: data.get('first_name'),
			last_name: data.get('last_name'),
			email: data.get('email'),
			password: data.get('password'),
			password_confirmation: data.get('password_confirmation')
		});

        console.log(response);
		if (!response.success) {
			return fail(401, response);
		}

		// const value = btoa(JSON.stringify(body.user));
		// cookies.set('jwt', value, { path: '/' });

		throw redirect(307, '/login');
	}
};
