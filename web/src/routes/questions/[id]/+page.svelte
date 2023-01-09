<script>
	import { enhance } from '$app/forms';
	import { page } from '$app/stores';

  /** @type {import('./$types').PageData} */
	export let data;
</script>

<svelte:head>
	<title>Details • Quizbox</title>
</svelte:head>

<div class="question">
  <div class="row">
    <h2>{data.question.title}</h2>
    <p><small>Asked by {data.question.user_id}</small></p>
    <div>
      {data.question.body}
    </div>
    <p><small><a href="/questions/{data.question.id}/vote?act=up&kind=question">Upvote (0)</a> | <a href="/questions/{data.question.id}/vote?act=down&kind=question">Downvote (0)</a></small></p>
  </div>
</div>

<hr />
<div class="row">
  <div class="col-md-8 col-md-offset-2">
    <h4>Answers</h4>
    {#if $page.data.user}
    <form use:enhance method="POST" action="?/createAnswer">
      <div class="form-group">
        <textarea class="form-control" name="body" id="body" cols="10" rows="5" placeholder="Answer"></textarea>
      </div>
      <button type="submit" class="btn btn-primary">Post your Answer</button>
    </form>
    <hr />
    {/if}
    <div class="answers">
      {#if data.answers.length > 0 }
        {#each data.answers as answer (answer.id)}
          <div class="answer">
              <div>{answer.body}</div>
              <small>
                By {answer.user_id} answered {answer.created_at} <a href={null}>Upvote</a> | <a href={null}>Downvote</a>
            </small>
            <hr />
          </div>
        {/each}
      {:else}
          <p>No Answer found.</p>
      {/if}

    </div>
  </div>
</div>
