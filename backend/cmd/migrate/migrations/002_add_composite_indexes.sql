-- +goose Up
CREATE INDEX idx_messages_channel_id_id ON messages (channel_id, id);
CREATE INDEX idx_dm_messages_dm_channel_id_id ON dm_messages (dm_channel_id, id);
CREATE INDEX idx_msg_reactions_message_user ON message_reactions (message_id, user_id);
CREATE INDEX idx_dm_msg_reactions_dm_message_user ON dm_message_reactions (dm_message_id, user_id);

-- +goose Down
DROP INDEX idx_messages_channel_id_id ON messages;
DROP INDEX idx_dm_messages_dm_channel_id_id ON dm_messages;
DROP INDEX idx_msg_reactions_message_user ON message_reactions;
DROP INDEX idx_dm_msg_reactions_dm_message_user ON dm_message_reactions;
