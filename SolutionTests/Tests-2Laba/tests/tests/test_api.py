# import pytest

# from lib.api_helper import *


# class TestHttpBaseChecks:
#     def tests_http_ok_on_root_page(self):
#         assert get_page().status_code == 200

#     def tests_http_text_non_empty(self):
#         assert get_page().text

#     def tests_http_ok_on_request_page(self):
#         assert post_form({'m':5, 'n':5}).status_code == 200

#     def tests_form_content_type_text_html(self):
#         assert get_page().headers['Content-Type'] == 'text/html; charset=utf-8'

#     def tests_form_content_type_app_json(self):
#         assert post_form({'m':5, 'n':5}).headers['Content-Type'] == 'application/json'


# class TestUniquePaths:

#     data_valid = {(3, 7), (3, 2),(1, 1), (5, 5),(3, 8), (15, 12),(10, 10)}

#     data_m_or_n_negative_or_0 = {(-1, 5), (0, 4), (20, -8), (2, 0), (0, 0), (-6, -12)}

#     data_m_or_n_not_integers = {(0.5,7), (8,11.2), (10.0,4),
#             ('!', 1), ('@', 1), ('#', 1), ('%', 1), (2, '$'),
#             (2, '!'), (2, '@'), (2, '#'), (2, '%'), ('$', 2),
#             ('@', '@'),('*', '*'),
#             ('а́', 5), ('©', 10), (20, '>'), ('🔥', '🔥'), ('◀️','▶️'), ('α', 4), (5/3, 3), (4, 6/2)}

#     data_m_and_n_invalid_borders = {(0, 10), (0, 101), (15, 101)}

#     data_m_and_n_valid_borders = {(1, 10), (2, 15), (1, 100), (5, 99), (1000, 1), (100000, 1), (20000000, 1)}

#     data_big_answer = {(50, 50), (12, 70), (100, 99), (30, 30), (20, 20)}

#     @pytest.mark.parametrize('data', data_valid)
#     def tests_true_if_valid_paths(self, data):
#         assert post_form({'m':data[0], 'n':data[1]}).json()['result']

#     @pytest.mark.parametrize('data', data_m_or_n_negative_or_0)
#     def test_m_or_n_negative_or_0(self, data):
#         assert post_form({'m': data[0], 'n': data[1]}).status_code == 400

#     @pytest.mark.parametrize('data', data_m_or_n_not_integers)
#     def test_m_or_n_not_integers(self, data):
#         assert post_form({'m': data[0], 'n': data[1]}).status_code == 400

#     @pytest.mark.parametrize('data', data_m_and_n_invalid_borders)
#     def test_m_and_n_invalid_borders(self, data):
#         assert post_form({'m': data[0], 'n': data[1]}).status_code == 400

#     @pytest.mark.parametrize('data', data_m_and_n_valid_borders)
#     def test_m_and_n_valid_borders(self, data):
#         assert post_form({'m':data[0], 'n':data[1]}).json()['result']

#     @pytest.mark.parametrize('data', data_big_answer)
#     def test_big_answer(self, data):
#         assert post_form({'m': data[0], 'n': data[1]}).status_code == 400

#     def test_bad_request_wrong_form_key(self):
#         assert post_form({'m': "testm", 'n': "testn"}).status_code == 400

#     def test_bad_request_no_data(self):
#         assert post_form({}).status_code == 500

#     def test_request_json(self):
#         assert post_form(None, json={'m': 3, 'n': 4}).status_code == 500