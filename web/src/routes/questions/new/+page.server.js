import { fail, redirect } from '@sveltejs/kit';
import * as api from '$lib/api.js';

/** @type {import('./$types').PageServerLoad} */
export async function load({ parent }) {
	const { user } = await parent();
	if (!user) throw redirect(307, '/');
}

/** @type {import('./$types').Actions} */
export const actions = {
	default: async ({ locals, request }) => {
		if (!locals.user) throw error(401);

    const data = await request.formData();
		const response = await api.post('questions', {
			user_id: locals.user.user.id,
			title: data.get('title'),
			body: data.get('body'),
			tags: data.get('tags'),
		}, locals.user.token);

		if (!response.success) {
			return fail(401, response);
		}

		throw redirect(307, '/');
	}
};
