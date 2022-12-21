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
