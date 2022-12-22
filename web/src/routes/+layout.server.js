import * as api from '$lib/api.js';

/** @type {import('./$types').LayoutServerLoad} */
export function load({ locals }) {
	return {
		user: locals.user && {
			id: locals.user.user.id,
			first_name: locals.user.user.first_name,
			last_name: locals.user.user.last_name,
			email: locals.user.email,
		}
	};
}

/** @type {import('./$types').Actions} */
export const actions = {
	logout: async ({ cookies, locals }) => {
		if (!locals.user) throw redirect(307, '/login');

    const response = await api.del('users/logout', locals.user.token);

    if (!response.success) {
			return fail(401, response.message);
    }

		cookies.delete('jwt', { path: '/' });
		locals.user = null;
		throw redirect(307, '/');
	},
};
