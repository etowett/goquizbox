import * as api from '$lib/api.js';

/** @type {import('./$types').PageServerLoad} */
export async function load({ locals, params }) {
	const [question, answers] = await Promise.all([
		api.get(`questions/${params.id}`, locals.user?.token),
		api.get(`questions/${params.id}/answers`, locals.user?.token)
	]);

	return {
		question: question.data,
		answers: answers.data.answers,
		pagination: answers.data.pagination,
	 };
}

/** @type {import('./$types').Actions} */
export const actions = {
	createAnswer: async ({ locals, params, request }) => {
		if (!locals.user) throw error(401);

		const data = await request.formData();
		await api.post(
			`questions/${params.id}/answers`,
			{
				user_id: locals.user.user.id,
				question_id: parseInt(params.id),
				body: data.get('body'),
			},
			locals.user.token
		);
	},
};
