import { fail, redirect } from '@sveltejs/kit';
import * as api from '$lib/api.js';

/** @type {import('./$types').PageServerLoad} */
export async function load({ locals }) {
	// if (locals.user) {
	// 	cookies.delete('jwt');
	// }
	if (locals.user) throw redirect(307, '/');
}

/** @type {import('./$types').Actions} */
export const actions = {
	default: async ({ cookies, request }) => {
		const data = await request.formData();

		const response = await api.post('users/login', {
      email: data.get('email'),
      password: data.get('password'),
      remember: data.get('remember') === 'on',
		});

		if (!response.success) {
			return fail(401, response);
		}

		const value = btoa(JSON.stringify(response));
		cookies.set('jwt', value, { path: '/' });

		throw redirect(307, '/');
	}
};
