import unittest
from unittest.mock import patch
import web_dev

class DevelopmentCleanup(unittest.TestCase):
    def test_readiness_failure_restores_static_proxy(self):
        with patch.object(web_dev,'restart',side_effect=[TimeoutError('readiness'),None]) as restart:
            with self.assertRaises(TimeoutError):
                with web_dev.development_proxy(['docker','compose']):
                    self.fail('failed startup yielded')
            self.assertEqual(restart.call_count,2)
            self.assertEqual(restart.call_args_list[-1].args,(['docker','compose'],'web'))

if __name__=='__main__':unittest.main()
