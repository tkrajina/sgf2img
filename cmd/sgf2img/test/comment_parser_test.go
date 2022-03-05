package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommentParser(t *testing.T) {
	t.Parallel()

	{
		cm := parseComment(`!img aaa l`, 19)
		assert.Equal(t, 1, len(cm.images))
		assert.Equal(t, 19/2, cm.images[0].right)
	}
	{
		cm := parseComment(`!img aaa r`, 19)
		assert.Equal(t, 1, len(cm.images))
		assert.Equal(t, 19/2, cm.images[0].left)
	}
	{
		cm := parseComment(`!img aaa u`, 19)
		assert.Equal(t, 1, len(cm.images))
		assert.Equal(t, 19/2, cm.images[0].down)
	}
	{
		cm := parseComment(`!img aaa d`, 19)
		assert.Equal(t, 1, len(cm.images))
		assert.Equal(t, 19/2, cm.images[0].up)
	}
	{
		cm := parseComment(`!img aaa 5l`, 19)
		assert.Equal(t, 1, len(cm.images))
		assert.Equal(t, 5, cm.images[0].right)
	}
	{
		cm := parseComment(`!img aaa 5r`, 19)
		assert.Equal(t, 1, len(cm.images))
		assert.Equal(t, 5, cm.images[0].left)
	}
	{
		cm := parseComment(`!img aaa 5u`, 19)
		assert.Equal(t, 1, len(cm.images))
		assert.Equal(t, 5, cm.images[0].down)
	}
	{
		cm := parseComment(`!img aaa 5d`, 19)
		assert.Equal(t, 1, len(cm.images))
		assert.Equal(t, 5, cm.images[0].up)
	}
}
