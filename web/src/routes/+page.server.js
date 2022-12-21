import * as api from '$lib/api.js';

export async function load() {
	const { data } = await api.get('questions');

	return {
		questions: data.questions,
		pagination: data.pagination
	};
}
