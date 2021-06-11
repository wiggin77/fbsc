# Focalboard Simple|Simulated Client

This app generates boards, cards, commands, and other block types within the Focalboard application. Viewing of boards, cards and views is also simulated. Any (reasonable) number of users can be simulated, with each user generating and retrieving blocks based on a probabilities defined in the configuration file.

Content for titles, descriptions and comments is `Lorem ipsum` generated text with randomized words, sentences, and paragraphs.

## Usage

```bash
./fbsc -f config.json
```

## Configuration

Create a default config:

```bash
./fbsc -c -f config.json
```

Modify the generated config file, at minimum adding a siteURL, and admin credentials. Workspaces, boards, cards and users will be created if needed.

config.json:

```json
{
  "site_url": "http://192.168.1.78:8065",
  "admin_username": "admin",
  "admin_password": "password",
  "workspaces": "",
  "user_names": [
    "darth-vader",
    "capt-steve"
  ],
  "user_count": 5,
  "boards": [
  ],
  "avg_action_delay": 15000,
  "delay_variance": 0.5,
  "prob_comment": 0.1,
  "prob_property": 0.01,
  "prob_description": 0.01,
  "prob_edit": 0.05,
  "prob_delete": 0.01,
  "max_words_per_sentence": 100,
  "max_sentences_per_paragraph": 20,
  "max_paragraphs_per_post": 2
}
```

| Field | Description |
| ----- | ----------- |
| site_url | Fully qualified URL to your Mattermost instance. |
| admin_username | Username of admin user for creating teams, channels, users. |
| admin_password | Password of admin user. |
| team_name | Name of team for posting. Will be created if needed. Users will be added to this team. |
| user_names | Array of user names to post as. Will be created and added to teams and channels if needed. |
| user_count | Number of additional randomly named users to add. |
| channel_names | Array of channels to post to. At least one is required. Will be created if needed. |
| avg_post_delay | Average time to wait between posts. Milliseconds. Actual delay is random. |
| delay_variance | Determines how much the random delay can vary from average as a percentage of `avg_post_delay`. (Range 0.0 - 1.0) |
| prob_react | Probability, as a precentage, that a user will react to a post. (Range 0.0 - 1.0) |
| prob_reply | Probability, as a precentage, that a user will reply to a post. (Range 0.0 - 1.0) |
| prob_edit | Probability, as a precentage, that a user will edit their own post. (Range 0.0 - 1.0) |
| prob_delete | Probability, as a precentage, that a user will delete their own post. (Range 0.0 - 1.0) |
| max_words_per_sentence | Maximum number of words in each sentence for randomly generated post text. |
| max_sentences_per_paragraph | Maximum number of sentences in each paragraph for randomly generated post text. |
| max_paragraphs_per_post | Maximum number of paragraphs in each randomly generated post. |
