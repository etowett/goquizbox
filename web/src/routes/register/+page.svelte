<script>
	import { enhance } from '$app/forms';
  import ListErrors from '$lib/ListErrors.svelte';

	/** @type {import('./$types').ActionData} */
	export let form;
</script>

<svelte:head>
	<title>Register • Quizbox</title>
</svelte:head>

<div class="row">
  <div class="col-md-6 col-offset-md-2">
    <h1 class="text-xs-center">Login In</h1>
    <p class="text-xs-center">
			<a href="/login">Have an account?</a>
    </p>

    <ListErrors errors={form?.errors} />

    <form use:enhance method="POST">
      <div class="form-group">
        <label for="first_name">First Name</label>
        <input type="text" class="form-control" name="first_name" id="first_name" placeholder="Enter First Name">
      </div>
      <div class="form-group">
        <label for="last_name">Last Name</label>
        <input type="text" class="form-control" name="last_name" id="last_name" placeholder="Enter Last Name">
      </div>
      <div class="form-group">
        <label for="email">Email address</label>
        <input type="email" class="form-control" name="email" id="email" placeholder="Enter email">
      </div>
      <div class="form-group">
        <label for="password">Password</label>
        <input type="password" class="form-control" name="password" id="password" placeholder="Password">
      </div>
      <div class="form-group">
        <label for="password_confirmation">Confirm Password</label>
        <input type="password" class="form-control" name="password_confirmation" id="password_confirmation" placeholder="Password Confirmation">
      </div>
      <button type="submit" class="btn btn-primary">Register</button>
    </form>
  </div>
</div>
