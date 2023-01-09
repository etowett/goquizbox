<script>
	import { enhance } from '$app/forms';
    import ListErrors from '$lib/ListErrors.svelte';

	/** @type {import('./$types').ActionData} */
	export let form;
</script>

<svelte:head>
	<title>Sign in • Quizbox</title>
</svelte:head>

<div class="row">
  <div class="col-md-6 col-offset-md-2">
    <h1 class="text-xs-center">Login In</h1>
    <p class="text-xs-center">
        <a href="/register">Need an account?</a>
    </p>

    <ListErrors errors={form?.errors} />

      <form use:enhance method="POST">
        <div class="form-group">
        <label for="email">Email address</label>
        <input type="email" class="form-control" name="email" id="email" placeholder="Enter email">
        </div>
        <div class="form-group">
          <label for="password">Password</label>
          <input type="password" class="form-control" name="password" id="password" placeholder="Password">
        </div>
        <div class="form-check">
          <input type="checkbox" class="form-check-input" name="remember" id="remember">
          <label class="form-check-label" for="remember">Remember</label>
        </div>
        <button type="submit" class="btn btn-primary">Login</button>
      </form>
  </div>
</div>
