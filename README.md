# Focalboard Simple|Simulated Client

This app generates boards, cards, content, and other block types within the Focalboard application. Any (reasonable) number of users can be simulated, with each user generating blocks based on limits defined in the configuration file.

Content for card text is `Lorem ipsum` generated text with randomized words, sentences, and paragraphs.

## Usage

```bash
./fbsc -f config.json
```

## Configuration

Create a default config:

```bash
./fbsc -c -f config.json
```

Modify the generated config file, at minimum adding a siteURL, a team name, and admin credentials. Team, workspaces, boards, cards and users will be created if needed.

config.json:

```json
{
  "site_url": "http://192.168.1.78:8065",
  "admin_username": "admin",
  "admin_password": "password",
  "team_name": "",
  "user_count": 10,
  "channels_per_user": 5,
  "boards_per_channel": 7,
  "cards_per_board": 10,
  "max_words_per_sentence": 100,
  "max_sentences_per_paragraph": 20,
  "max_paragraphs_per_comment": 2,
  "board_delay_ms": 10,
  "card_delay_ms": 10
}
```

| Field | Description |
| ----- | ----------- |
| site_url | Fully qualified URL to your Mattermost instance. |
| admin_username | Username of admin user for creating workspaces and users. |
| admin_password | Password of admin user. |
| team_name |  Name of team add workspaces and boards to. Will be created if it does not exist. |
| user_count | Number of users to create. |
| channels_per_user | Number of channels (workspaces) to create for each user. |
| boards_per_channel | Number of boards to create for each channel (workspace). |
| cards_per_board | Number of cards to create for each board. |
| max_words_per_sentence | Maximum number of words in each sentence for randomly generated card text. |
| max_sentences_per_paragraph | Maximum number of sentences in each paragraph for randomly generated card text. |
| max_paragraphs_per_comment | Maximum number of paragraphs in each card description. |
| board_delay_ms | Number of milliseconds to sleep after creating a board. Use this to throttle during load testing. |
| card_delay_ms | Number of milliseconds to sleep after creating a board. Use this to throttle during load testing. |
