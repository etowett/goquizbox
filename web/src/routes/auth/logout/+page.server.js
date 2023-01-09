import { fail, redirect } from '@sveltejs/kit';
import * as api from '$lib/api.js';

/** @type {import('./$types').PageServerLoad} */
export async function load({ locals, cookies }) {
  if (!locals.user) throw redirect(307, '/auth/login');

  const response = await api.del('auth/logout', locals.user.token);

	if (!response.success) {
		return fail(401, response.message);
	}

	cookies.delete('jwt', { path: '/' });
	locals.user = null;
	throw redirect(307, '/auth/login');
}
