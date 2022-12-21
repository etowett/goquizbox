<script>
	import { enhance } from '$app/forms';
  import ListErrors from '$lib/ListErrors.svelte';
  import { page } from '$app/stores';

	/** @type {import('./$types').ActionData} */
	export let form;
</script>

<svelte:head>
	<title>New Question • Quizbox</title>
</svelte:head>

<h2>Add Question</h2>

<div class="row">
  <div class="col-md-8 col-offset-md-2">
    <ListErrors errors={form?.errors} />

    <form use:enhance method="POST">
      <div class="form-group">
        <label for="title">Title</label>
        <input type="text" class="form-control" name="title" id="title" placeholder="Title">
      </div>
      <div class="form-group">
        <label for="body">Body</label>
        <textarea class="form-control" name="body" id="body" cols="30" rows="10" placeholder="Body"></textarea>
      </div>
      <div class="form-group">
        <label for="tags">Tags</label>
        <input type="text" class="form-control" name="tags" id="tags" placeholder="Tags">
      </div>
      <button type="submit" class="btn btn-primary">Post your Question</button>
    </form>
  </div>
</div>
