# AI instruction 49

## Password security

Here's the current policy of CWCloud:

```python
TEST_PASSWORDS = {
    'valid': {
    'value': 'ValidPass123$',
    'expected_status': True,
    'expected_code': None
    },
    'no_capital': {
        'value': 'lowercase123$',
        'expected_status': False,
        'expected_code': 'password_no_upper'
    },
    'no_lower': {
        'value': 'UPPERCASE123$',
        'expected_status': False,
        'expected_code': 'password_no_lower'
    },
    'no_special': {
        'value': 'NoSpecial123',
        'expected_status': False,
        'expected_code': 'password_no_symbol'
    },
    'too_short': {
        'value': 'Short1$',
        'expected_status': False,
        'expected_code': 'password_too_short'
    }
}
```

I want the same policy with an utils function `utils.IsPasswordValid` which is also giving an error code to be sent through `i18n_code` (apply on every form changing the password: signup, forgotten password, settings, etc).
