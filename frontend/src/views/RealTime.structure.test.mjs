import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { parse as parseSfc } from '@vue/compiler-sfc'
import { NodeTypes, parse as parseTemplate } from '@vue/compiler-dom'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const viewPath = path.join(__dirname, 'RealTime.vue')

async function loadTemplateAst() {
    const source = await readFile(viewPath, 'utf8')
    const { descriptor, errors } = parseSfc(source, { filename: viewPath })

    assert.equal(errors.length, 0, `RealTime.vue should parse cleanly: ${errors.join(', ')}`)
    assert.ok(descriptor.template?.content, 'RealTime.vue should contain a template block')

    return parseTemplate(descriptor.template.content, { comments: false })
}

function walk(node, visit) {
    visit(node)

    if (Array.isArray(node.children)) {
        for (const child of node.children) {
            walk(child, visit)
        }
    }

    if (node.branches) {
        for (const branch of node.branches) {
            walk(branch, visit)
        }
    }
}

function getElements(ast) {
    const elements = []
    walk(ast, (node) => {
        if (node.type === NodeTypes.ELEMENT) {
            elements.push(node)
        }
    })
    return elements
}

function getAttributeValue(element, name) {
    const attribute = element.props.find(
        (prop) => prop.type === NodeTypes.ATTRIBUTE && prop.name === name
    )

    return attribute?.value?.content ?? null
}

function hasClassToken(element, className) {
    const classValue = getAttributeValue(element, 'class')
    return classValue
        ? classValue.split(/\s+/).some((token) => token === className)
        : false
}

function hasDirective(element, name, argumentName, expressionPattern, modifier) {
    return element.props.some((prop) => {
        if (prop.type !== NodeTypes.DIRECTIVE || prop.name !== name) {
            return false
        }

        if ((argumentName ?? null) !== (prop.arg?.content ?? null)) {
            return false
        }

        if (modifier) {
            const modifiers = prop.modifiers.map((entry) =>
                typeof entry === 'string' ? entry : entry?.content
            )

            if (!modifiers.includes(modifier)) {
                return false
            }
        }

        if (!expressionPattern) {
            return true
        }

        return expressionPattern.test(prop.exp?.content ?? '')
    })
}

test('RealTime template removes inline history list rendering', async () => {
    const elements = getElements(await loadTemplateAst())
    const rendersInlineHistory = elements.some((element) =>
        hasDirective(element, 'bind', 'data-source', /^\s*user\.history\s*$/)
    )

    assert.equal(
        rendersInlineHistory,
        false,
        'expected inline user.history list rendering to be removed from the loading card template'
    )
})

test('RealTime template keeps the history drawer and dedicated history action path', async () => {
    const elements = getElements(await loadTemplateAst())

    const hasHistoryDrawer = elements.some(
        (element) =>
            element.tag === 'a-drawer' &&
            hasDirective(element, 'model', 'open', /^\s*historyDrawerOpen\s*$/)
    )
    assert.equal(hasHistoryDrawer, true, 'expected the history drawer to remain wired to historyDrawerOpen')

    const keepsGameHistoryList = elements.some(
        (element) =>
            element.tag === 'GameHistoryList' &&
            hasDirective(element, 'bind', 'page-size', /^\s*20\s*$/)
    )
    assert.equal(keepsGameHistoryList, true, 'expected GameHistoryList to remain with page-size 20')

    const hasDedicatedHistoryAction = elements.some(
        (element) =>
            hasClassToken(element, 'history-action') &&
            hasDirective(element, 'on', 'click', /^\s*openPlayerHistory\(user\)\s*$/, 'stop')
    )
    assert.equal(
        hasDedicatedHistoryAction,
        true,
        'expected a dedicated history action element wired to openPlayerHistory(user)'
    )
})

test('RealTime template keeps the planned loading-card structure hooks', async () => {
    const elements = getElements(await loadTemplateAst())

    for (const className of ['match-board', 'team-row', 'player-card', 'player-nameplate']) {
        assert.equal(
            elements.some((element) => hasClassToken(element, className)),
            true,
            `expected template to include a .${className} structural hook`
        )
    }
})
